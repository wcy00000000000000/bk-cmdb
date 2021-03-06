/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package synchronizer

import (
	"configcenter/src/common/blog"
	"configcenter/src/scene_server/admin_server/synchronizer/meta"
)

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id int, workerQueue chan meta.WorkRequest, handler meta.SyncHandler) *Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		SyncHandler: handler,
	}

	return &worker
}

// Worker represent a worker
type Worker struct {
	ID          int
	WorkerQueue chan meta.WorkRequest
	QuitChan    chan bool
	SyncHandler meta.SyncHandler
}

// Start "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		for {
			select {
			case work := <-w.WorkerQueue:
				// Receive a work request.
				blog.Infof("worker%d: Received work request, delaying for %f seconds\n", w.ID, work.Delay.Seconds())
				w.doWork(&work)

			case <-w.QuitChan:
				// We have been asked to stop.
				blog.Infof("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}

func (w *Worker) doWork(work *meta.WorkRequest) error {
	blog.Infof("start doing work: %+v", work)
	switch work.ResourceType {
	case meta.BusinessResource:
		w.SyncHandler.HandleBusinessSync(work)
	case meta.HostResource:
		w.SyncHandler.HandleHostSync(work)
	case meta.SetResource:
		w.SyncHandler.HandleSetSync(work)
	case meta.ModuleResource:
		w.SyncHandler.HandleModuleSync(work)
	case meta.ModelResource:
		w.SyncHandler.HandleModelSync(work)
	case meta.InstanceResource:
		w.SyncHandler.HandleInstanceSync(work)
	case meta.AuditCategory:
		w.SyncHandler.HandleAuditSync(work)
	case meta.ProcessResource:
		w.SyncHandler.HandleProcessSync(work)
	case meta.DynamicGroupResource:
		w.SyncHandler.HandleDynamicGroupSync(work)
	case meta.ClassificationResource:
		w.SyncHandler.HandleClassificationSync(work)

	default:
		blog.Errorf("work type:%s didn't register yet.", work.ResourceType)

	}
	return nil
}
