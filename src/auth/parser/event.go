/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parser

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"configcenter/src/auth/meta"
)

func (ps *parseStream) eventRelated() *parseStream {
	if ps.shouldReturn() {
		return ps
	}

	ps.subscribe()

	return ps
}

var (
	findSubscribePattern   = regexp.MustCompile(`^/api/v3/event/subscribe/search/\S+/\d+/?$`)
	createSubscribePattern = regexp.MustCompile(`^/api/v3/event/subscribe/\S+/\d+/?$`)
	updateSubscribePattern = regexp.MustCompile(`^/api/v3/event/subscribe/\S+/\d+/\d+/?$`)
	deleteSubscribePattern = regexp.MustCompile(`^/api/v3/event/subscribe/\S+/\d+/\d+/?$`)
)

func (ps *parseStream) subscribe() *parseStream {
	if ps.shouldReturn() {
		return ps
	}

	// find all the subscription
	if ps.hitRegexp(findSubscribePattern, http.MethodPost) {
		ps.Attribute.Resources = []meta.ResourceAttribute{
			meta.ResourceAttribute{
				Basic: meta.Basic{
					Type:   meta.EventPushing,
					Action: meta.FindMany,
				},
			},
		}
		return ps
	}

	// create a subscription
	if ps.hitRegexp(createSubscribePattern, http.MethodPost) {
		ps.Attribute.Resources = []meta.ResourceAttribute{
			meta.ResourceAttribute{
				Basic: meta.Basic{
					Type:   meta.EventPushing,
					Action: meta.Create,
				},
			},
		}
		return ps
	}

	// update a subscription
	if ps.hitRegexp(updateSubscribePattern, http.MethodPut) {
		bizID, err := strconv.ParseInt(ps.RequestCtx.Elements[6], 10, 64)
		if err != nil {
			ps.err = fmt.Errorf("update subscription batch, but got invalid subscription id: %s", ps.RequestCtx.Elements[4])
			return ps
		}
		ps.Attribute.Resources = []meta.ResourceAttribute{
			meta.ResourceAttribute{
				Basic: meta.Basic{
					Type:       meta.EventPushing,
					Action:     meta.Update,
					InstanceID: bizID,
				},
			},
		}
		return ps
	}

	// delete a subscription
	if ps.hitRegexp(deleteSubscribePattern, http.MethodDelete) {
		subscribeID, err := strconv.ParseInt(ps.RequestCtx.Elements[6], 10, 64)
		if err != nil {
			ps.err = fmt.Errorf("update subscription batch, but got invalid subscription id: %s", ps.RequestCtx.Elements[4])
			return ps
		}
		ps.Attribute.Resources = []meta.ResourceAttribute{
			meta.ResourceAttribute{
				Basic: meta.Basic{
					Type:       meta.EventPushing,
					Action:     meta.Delete,
					InstanceID: subscribeID,
				},
			},
		}
		return ps
	}

	return ps
}
