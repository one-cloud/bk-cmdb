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

package operation

import (
	"context"

	"configcenter/src/apimachinery"
	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/scene_server/topo_server/core/types"
)

type AuditOperationInterface interface {
	Query(params types.ContextParams, query *metadata.QueryInput) (interface{}, error)
}

// NewAuditOperation create a new inst operation instance
func NewAuditOperation(client apimachinery.ClientSetInterface) AuditOperationInterface {
	return &audit{
		clientSet: client,
	}
}

type audit struct {
	clientSet apimachinery.ClientSetInterface
}

func (a *audit) TranslateOpLanguage(params types.ContextParams, input interface{}) mapstr.MapStr {

	data, err := mapstr.NewFromInterface(input)
	if nil != err {
		blog.Errorf("translate failed, err: %+v", err)
		return data
	}

	info, err := data.MapStrArray("info")
	if nil != err {
		return data
	}

	for _, row := range info {

		opDesc, err := row.String(common.BKOpDescField)
		if nil != err {
			continue
		}
		newDesc := params.Lang.Language("auditlog_" + opDesc)
		if "" == newDesc {
			continue
		}
		row.Set(common.BKOpDescField, newDesc)
	}
	return data
}

func (a *audit) Query(params types.ContextParams, query *metadata.QueryInput) (interface{}, error) {
	rsp, err := a.clientSet.AuditController().GetAuditLog(context.Background(), params.Header, query)
	if nil != err {
		blog.Errorf("[audit] failed request audit controller, error info is %s", err.Error())
		return nil, params.Err.New(common.CCErrCommHTTPDoRequestFailed, err.Error())
	}

	if !rsp.Result {
		blog.Errorf("[audit] failed request audit controller, error info is %s", rsp.ErrMsg)
		return nil, params.Err.New(common.CCErrAuditTakeSnapshotFaile, rsp.ErrMsg)
	}

	return a.TranslateOpLanguage(params, rsp.Data), nil
}
