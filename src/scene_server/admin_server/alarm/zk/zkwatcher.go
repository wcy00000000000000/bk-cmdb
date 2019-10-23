package zk

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"configcenter/src/common/blog"
	"configcenter/src/common/types"
	"configcenter/src/common/zkclient"
)

type ZkWatcherEventHandler interface {
	OnAddLeaf(branch, leaf, value string)
	OnDeleteLeaf(branch, leaf, value string, errorAlarm bool)
}

func NewZkWatcher(zkAddrs []string, baseBranch string, callBacker ZkWatcherEventHandler) (*ZkWatcher, error) {
	cli := zkclient.NewZkClient(zkAddrs)
	if err := cli.ConnectEx(3 * time.Second); nil != err {
		return nil, fmt.Errorf("new zookeeper client failed. err: %v", err)
	}
	return &ZkWatcher{
		zkCli:      cli,
		BaseBranch: baseBranch,
		CallBacker: callBacker,
		resourceBox: &Box{
			Resource: make(map[string]map[string]string),
		},
	}, nil
}

type Box struct {
	Lock     sync.Mutex
	Resource map[string]map[string]string
}

type ZkWatcher struct {
	zkCli       *zkclient.ZkClient
	BaseBranch  string
	CallBacker  ZkWatcherEventHandler
	resourceBox *Box
}

func (z *ZkWatcher) Run() error {
	if len(z.BaseBranch) == 0 {
		return errors.New("base branch can not be null")
	}

	if nil == z.CallBacker {
		return errors.New("event call backer can not be nil")
	}

	if err := z.TryWatchBranch(z.BaseBranch); nil != err {
		return err
	}

	return nil
}

func (z *ZkWatcher) TryWatchBranch(zkBranch string) error {
	if true == z.IsExist(zkBranch) {
		return nil
	}
	// a new branch, add this branch first.
	if zkBranch != z.BaseBranch {
		z.AddNewBranch(zkBranch)
	}

	blog.Infof("start to watch new branch %s", zkBranch)

	go func() {
		for {
			childrens, eventChan, err := z.zkCli.WatchChildren(zkBranch)
			if err != nil {
				blog.Errorf("watch zookeeper children %s failed. err: %v", zkBranch, err)
				if strings.Contains(err.Error(), "node does not exist") == true {
					z.DeleteBranch(zkBranch)
					blog.Errorf("zookeeper node:%s does not exist. stop watch and close connection now.", zkBranch)
					return
				}
				time.Sleep(10 * time.Second)
				continue
			}
			leafMapper := make(map[string]struct{})
			for _, child := range childrens {
				if true == z.IsLeaf(child) {
					leafMapper[child] = struct{}{}
					path := fmt.Sprintf("%s/%s", zkBranch, child)
					if err := z.UpdateLeaf(z.zkCli, zkBranch, child); nil != err {
						blog.Errorf("update leaf failed. path: %v, err: %v", path, err)
						continue
					}
					continue
				}

				// this is a new branch
				newBranch := fmt.Sprintf("%s/%s", zkBranch, child)
				if err := z.TryWatchBranch(newBranch); nil != err {
					blog.Errorf("watch new branch: %s failed. err: %v", newBranch, err)
					continue
				}
			}
			// check leave child
			z.CheckRedundancyLeaf(zkBranch, leafMapper)

			// wait for another change.
			event := <-eventChan
			blog.Infof("path(%s) in zookeeper received an event, reason :%s.", zkBranch, event.Type)
			if event.Type == zkclient.EventNodeDeleted {
				z.DeleteBranch(zkBranch)
				blog.Infof("zookeeper node: %s is deleted, stop watch and close the connection.", zkBranch)
				return
			}
		}
	}()
	return nil
}

func (z *ZkWatcher) IsLeaf(child string) bool {
	return strings.HasPrefix(child, "_c_")
}

func (r *ZkWatcher) IsExist(key string) bool {
	r.resourceBox.Lock.Lock()
	defer r.resourceBox.Lock.Unlock()
	_, exist := r.resourceBox.Resource[key]
	return exist
}

func (z *ZkWatcher) CheckRedundancyLeaf(branch string, children map[string]struct{}) {
	z.resourceBox.Lock.Lock()
	defer z.resourceBox.Lock.Unlock()
	branchRes, exist := z.resourceBox.Resource[branch]
	if false == exist {
		blog.Warnf("check the redundancy leaf, but the branch: %s do not exist in the resource", branch)
		return
	}

	lost := make(map[string][]string)
	for subKey, value := range branchRes {
		if _, exist := children[subKey]; !exist {
			// find a redundancy leaf, need to alarm.
			blog.Warnf("find a redundancy leaf: %s in branch: %s, old value: %#v", subKey, branch, value)
			delete(z.resourceBox.Resource[branch], subKey)
			z.CallBacker.OnDeleteLeaf(branch, subKey, value, false)
			lost[branch] = append(lost[branch], value)
		}
	}

	_, exist = lost[branch]
	if !exist{
		return
	}

	b := branch
	time.AfterFunc(60*time.Second, func() {
		branch := b
		newChildren, _, err := z.zkCli.WatchChildren(branch)
		if err != nil {
			blog.Errorf("watch zookeeper children %s failed. err: %v", branch, err)
			if strings.Contains(err.Error(), "node does not exist") == true {
				z.DeleteBranch(branch)
				blog.Errorf("zookeeper node:%s does not exist. stop watch and close connection now.", branch)
				return
			}
		}
		leafMapper := make(map[string]string)
		for _, child := range newChildren {
			path := fmt.Sprintf("%s/%s", branch, child)
			value, err := z.zkCli.Get(path)
			if nil != err {
				blog.Errorf("get path: %s value failed. err: %v", path, err)
			}
			var srvInfo types.ServerInfo
			if err := json.Unmarshal([]byte(value), &srvInfo); nil != err {
				blog.Errorf("unmarshal leaf %s/%s failed. err: %v", branch, child, err)
				return
			}
			leafMapper[child] = srvInfo.Instance()
		}
		l, exist := lost[branch]
		if !exist{
			return
		}
		for _, value := range l {
			flag := false
			var srvInfo types.ServerInfo
			if err := json.Unmarshal([]byte(value), &srvInfo); nil != err {
				blog.Errorf("unmarshal value %s failed. err: %v", value, err)
				return
			}
			for _, v := range leafMapper {
				if srvInfo.Instance() == v {
					flag = true
					break
				}
			}
			if !flag{
				z.CallBacker.OnDeleteLeaf(branch, "", value, true)
			}
		}
	})
}

func (z *ZkWatcher) Set(key, subkey, value string) {
	z.resourceBox.Lock.Lock()
	defer z.resourceBox.Lock.Unlock()
	if _, exist := z.resourceBox.Resource[key]; false == exist {
		z.resourceBox.Resource[key] = make(map[string]string)
	}
	old, exist := z.resourceBox.Resource[key][subkey]
	if false == exist {
		// a new key
		z.resourceBox.Resource[key][subkey] = value
		blog.Infof("add new key: %s, value: %s", fmt.Sprintf("%s/%s", key, subkey), value)
		z.CallBacker.OnAddLeaf(key, subkey, value)
		return
	}

	if 0 == strings.Compare(old, value) {
		return
	}
	// need to update this key.
	z.resourceBox.Resource[key][subkey] = value
	blog.Infof("update changed key:%s, old: %s, new: %s", fmt.Sprintf("%s/%s", key, subkey), old, value)
	return
}

func (z *ZkWatcher) AddNewBranch(branch string) {
	z.resourceBox.Lock.Lock()
	defer z.resourceBox.Lock.Unlock()
	if _, exist := z.resourceBox.Resource[branch]; false == exist {
		z.resourceBox.Resource[branch] = make(map[string]string)
		return
	}
	return
}

func (z *ZkWatcher) DeleteBranch(branch string) {
	z.resourceBox.Lock.Lock()
	leafs, exist := z.resourceBox.Resource[branch]
	if false == exist {
		return
	}
	delete(z.resourceBox.Resource, branch)
	z.resourceBox.Lock.Unlock()

	blog.Warnf("delete branch: %s.", branch)

	time.AfterFunc(60*time.Second, func() {
		_, _, err := z.zkCli.WatchChildren(branch)
		if err != nil {
			blog.Errorf("watch zookeeper children %s failed. err: %v", branch, err)
			if strings.Contains(err.Error(), "node does not exist") == true {
				for leaf, srvinfo := range leafs {
					z.CallBacker.OnDeleteLeaf(branch, leaf, srvinfo, true)
				}
				return
			}
		}
	})

	for leaf, srvinfo := range leafs {
		z.CallBacker.OnDeleteLeaf(branch, leaf, srvinfo, false)
	}
}

func (z *ZkWatcher) UpdateLeaf(zkCli *zkclient.ZkClient, fatherPath, child string) error {
	path := fmt.Sprintf("%s/%s", fatherPath, child)
	value, err := zkCli.Get(path)
	if nil != err {
		return fmt.Errorf("get path: %s value failed. err: %v", path, err)
	}
	z.Set(fatherPath, child, string(value))
	return nil
}
