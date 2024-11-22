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

package watch

import (
	"errors"

	"configcenter/src/storage/stream/types"
)

// EventType TODO
type EventType string

const (
	// Create TODO
	Create EventType = "create"
	// Update TODO
	Update EventType = "update"
	// Delete TODO
	Delete EventType = "delete"
	// Unknown TODO
	Unknown EventType = "unknown"
)

// Validate TODO
func (e EventType) Validate() error {
	switch e {
	case Create, Update, Delete:
		return nil
	default:
		return errors.New("unsupported event type")
	}
}

// ConvertOperateType TODO
func ConvertOperateType(typ types.OperType) EventType {
	switch typ {
	case types.Insert:
		return Create
	case types.Replace, types.Update:
		return Update
	case types.Delete:
		return Delete
	default:
		return Unknown
	}
}

// ChainNode TODO
type ChainNode struct {
	// self increasing id, used for sequential batch query
	ID uint64 `json:"id" bson:"id"`
	// event's cluster time
	ClusterTime types.TimeStamp `json:"cluster_time" bson:"cluster_time"`
	// event's document object id, as is the value of "_id" in a document.
	Oid string `json:"oid" bson:"oid"`
	// event's type
	EventType EventType `json:"type" bson:"type"`
	// event's resume token, if token is "", then it's a head, and no tail
	Token string `json:"token" bson:"token"`
	// generated by cmdb, and is 1:1 with mongodb's resume token
	// the current node's cursor
	Cursor string `json:"cursor" bson:"cursor"`
	// InstanceID object instance's ID, preserved for latter event aggregation operation
	InstanceID int64 `json:"inst_id,omitempty" bson:"inst_id,omitempty"`
	// SubResource the sub resource of the watched resource, eg. the object ID of the instance resource
	SubResource []string `json:"bk_sub_resource,omitempty" bson:"bk_sub_resource,omitempty"`
	// TenantID the supplier account of the chain node's related event resource.
	TenantID string `json:"tenant_id" bson:"tenant_id"`
}

// LastChainNodeData TODO
type LastChainNodeData struct {
	Coll        string          `json:"_id" bson:"_id"`
	ID          uint64          `json:"id" bson:"id"`
	Token       string          `json:"token" bson:"token"`
	Cursor      string          `json:"cursor" bson:"cursor"`
	StartAtTime types.TimeStamp `json:"start_at_time,omitempty" bson:"start_at_time,omitempty"`
}
