/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/metadata"
	"configcenter/src/common/universalsql/mongo"
)

// HasInstance used to check if the model has some instances
func (s *coreService) HasInstance(kit *rest.Kit, objIDS []string) (exists bool, err error) {

	// TODO: need to implement a new query function which is used to count the instances for the all objIDS
	cond := new(metadata.Condition)
	for _, objID := range objIDS {
		results, err := s.core.InstanceOperation().CountModelInstances(kit, objID, cond)
		if err != nil {
			return false, err
		}
		if results.Count > 0 {
			return true, nil
		}
	}

	return false, nil
}

// HasAssociation used to check if the model has some associations
func (s *coreService) HasAssociation(kit *rest.Kit, objIDS []string) (exists bool, err error) {

	// construct the model association query condition
	cond := mongo.NewCondition()
	cond.Or(&mongo.In{Key: metadata.AssociationFieldObjectID, Val: objIDS})
	cond.Or(&mongo.In{Key: metadata.AssociationFieldAsstID, Val: objIDS})

	// check the model association
	countCond := &metadata.Condition{Condition: cond.ToMapStr()}
	result, err := s.core.AssociationOperation().CountModelAssociations(kit, countCond)
	if err != nil {
		return false, err
	}
	if result.Count > 0 {
		return true, nil
	}

	return false, nil
}

// CascadeDeleteAssociation cascade delete all associated data (included instances, model association, instance association) associated with modelObjID
func (s *coreService) CascadeDeleteAssociation(kit *rest.Kit, objIDS []string) error {

	// cascade delete the modelIDS
	if err := s.CascadeDeleteInstances(kit, objIDS); err != nil {
		return err
	}

	// construct the deletion command
	cond := mongo.NewCondition()
	cond.Element(&mongo.Eq{Key: common.TenantID, Val: kit.TenantID})
	cond.Or(&mongo.In{Key: metadata.AssociationFieldObjectID, Val: objIDS})
	cond.Or(&mongo.In{Key: metadata.AssociationFieldAssociationObjectID, Val: objIDS})

	// execute delete command
	_, err := s.core.AssociationOperation().CascadeDeleteModelAssociation(kit,
		metadata.DeleteOption{Condition: cond.ToMapStr()})
	if err != nil {
		blog.Errorf("aborted to cascade the model associations by the condition (%v), err: %v, rid: %s",
			cond.ToMapStr(), err, kit.Rid)
		return err
	}

	return err
}

// CascadeDeleteInstances cascade delete all instances associated with modelObjID
// (included instances, instance association)
func (s *coreService) CascadeDeleteInstances(kit *rest.Kit, objIDS []string) error {

	// construct the deletion command which is used to delete all instances
	for _, objID := range objIDS {
		_, err := s.core.InstanceOperation().CascadeDeleteModelInstance(kit, objID, metadata.DeleteOption{})
		if err != nil {
			blog.Errorf("aborted to cascade delete the association for the model objectID(%s), err: %v, rid: %s", objID,
				err, kit.Rid)
			return err
		}
	}

	return nil
}
