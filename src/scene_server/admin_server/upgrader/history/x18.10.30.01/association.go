// Package x18_10_30_01 TODO
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
package x18_10_30_01

import (
	"context"
	"fmt"
	"time"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/condition"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/scene_server/admin_server/upgrader/history"
	"configcenter/src/storage/dal"
	"configcenter/src/storage/dal/types"

	"go.mongodb.org/mongo-driver/bson"
)

func createAssociationTable(ctx context.Context, db dal.RDB, conf *history.Config) error {
	tablenames := []string{common.BKTableNameAsstDes, common.BKTableNameObjAsst, common.BKTableNameInstAsst}
	for _, tablename := range tablenames {
		exists, err := db.HasTable(ctx, tablename)
		if err != nil {
			return err
		}
		if !exists {
			if err = db.CreateTable(ctx, tablename); err != nil && !db.IsDuplicatedError(err) {
				return err
			}
		}
	}
	return nil
}

func createInstanceAssociationIndex(ctx context.Context, db dal.RDB, conf *history.Config) error {

	idxArr, err := db.Table(common.BKTableNameInstAsst).Indexes(ctx)
	if err != nil {
		blog.Errorf("get table %s index error, err: %v", common.BKTableNameInstAsst, err)
		return err
	}

	createIdxArr := []types.Index{
		{Name: "idx_id", Keys: bson.D{{"id", -1}}, Background: true, Unique: true},
		{Name: "idx_objID_asstObjID_asstID", Keys: bson.D{{"bk_obj_id", -1}, {"bk_asst_obj_id", -1},
			{"bk_asst_id", -1}}},
	}
	for _, idx := range createIdxArr {
		exist := false
		for _, existIdx := range idxArr {
			if existIdx.Name == idx.Name {
				exist = true
				break
			}
		}
		// index already exist, skip create
		if exist {
			continue
		}
		if err := db.Table(common.BKTableNameInstAsst).CreateIndex(ctx, idx); err != nil && !db.IsDuplicatedError(err) {
			blog.Errorf("create index to cc_InstAsst error, err: %v, current index: %+v, all create index: %+v",
				err, idx, createIdxArr)
			return err
		}

	}

	return nil

}

func addPresetAssociationType(ctx context.Context, db dal.RDB, conf *history.Config) error {
	tablename := common.BKTableNameAsstDes

	asstTypes := []AssociationKind{
		{
			AssociationKindID:       "belong",
			AssociationKindName:     "",
			OwnerID:                 conf.TenantID,
			SourceToDestinationNote: "属于",
			DestinationToSourceNote: "包含",
			Direction:               metadata.DestinationToSource,
			IsPre:                   ptrue(),
		},
		{
			AssociationKindID:       "group",
			AssociationKindName:     "",
			OwnerID:                 conf.TenantID,
			SourceToDestinationNote: "组成",
			DestinationToSourceNote: "组成于",
			Direction:               metadata.DestinationToSource,
			IsPre:                   ptrue(),
		},
		{
			AssociationKindID:       "bk_mainline",
			AssociationKindName:     "",
			OwnerID:                 conf.TenantID,
			SourceToDestinationNote: "组成",
			DestinationToSourceNote: "组成于",
			Direction:               metadata.DestinationToSource,
			IsPre:                   ptrue(),
		},
		{
			AssociationKindID:       "run",
			AssociationKindName:     "",
			OwnerID:                 conf.TenantID,
			SourceToDestinationNote: "运行于",
			DestinationToSourceNote: "运行",
			Direction:               metadata.DestinationToSource,
			IsPre:                   ptrue(),
		},
		{
			AssociationKindID:       "connect",
			AssociationKindName:     "",
			OwnerID:                 conf.TenantID,
			SourceToDestinationNote: "上联",
			DestinationToSourceNote: "下联",
			Direction:               metadata.DestinationToSource,
			IsPre:                   ptrue(),
		},
		{
			AssociationKindID:       "default",
			AssociationKindName:     "默认关联",
			OwnerID:                 conf.TenantID,
			SourceToDestinationNote: "关联",
			DestinationToSourceNote: "被关联",
			Direction:               metadata.DestinationToSource,
			IsPre:                   ptrue(),
		},
	}

	for _, asstType := range asstTypes {
		_, _, err := history.Upsert(ctx, db, tablename, asstType, "id", []string{"bk_asst_id"}, []string{"id"})
		if err != nil {
			return err
		}
	}
	return nil
}

// Association association struct
type Association struct {
	metadata.Association `bson:",inline"`
	ObjectAttID          string `bson:"bk_object_att_id"`
}

func reconcilAsstData(ctx context.Context, db dal.RDB, conf *history.Config) error {

	assts := []Association{}
	err := db.Table(common.BKTableNameObjAsst).Find(nil).All(ctx, &assts)
	if err != nil {
		return err
	}

	propertyCond := condition.CreateCondition()
	propertyCond.Field(common.BKPropertyTypeField).In([]string{"multiasst", "singleasst"})
	propertys := []metadata.ObjAttDes{}
	err = db.Table(common.BKTableNameObjAttDes).Find(propertyCond.ToMapStr()).All(ctx, &propertys)
	if err != nil {
		return err
	}

	properyMap := map[string]metadata.ObjAttDes{}
	buildObjPropertyMapKey := func(objID string, propertyID string) string {
		return fmt.Sprintf("%s:%s", objID, propertyID)
	}
	for _, property := range propertys {
		properyMap[buildObjPropertyMapKey(property.ObjectID, property.PropertyID)] = property
		blog.Infof("key %s: %+v", buildObjPropertyMapKey(property.ObjectID, property.PropertyID), property)
	}

	flag := "updateflag"
	for _, asst := range assts {
		if asst.ObjectAttID == "bk_childid" {
			asst.AsstKindID = common.AssociationKindMainline
			asst.AssociationName = buildObjAsstID(asst)
			asst.Mapping = metadata.OneToOneMapping
			asst.OnDelete = metadata.NoAction
			if (asst.ObjectID == common.BKInnerObjIDModule && asst.AsstObjID == common.BKInnerObjIDSet) ||
				(asst.ObjectID == common.BKInnerObjIDHost && asst.AsstObjID == common.BKInnerObjIDModule) {
				asst.IsPre = ptrue()
			} else {
				asst.IsPre = pfalse()
			}

			// update ObjAsst
			updateCond := condition.CreateCondition()
			updateCond.Field("id").Eq(asst.ID)
			if err = db.Table(common.BKTableNameObjAsst).Update(ctx, updateCond.ToMapStr(), asst); err != nil {
				return err
			}

			// update InstAsst
			updateInst := mapstr.New()
			updateInst.Set("bk_obj_asst_id", asst.AssociationName)
			updateInst.Set("bk_asst_id", asst.AsstKindID)
			updateInst.Set("last_time", time.Now())
			err = db.Table(common.BKTableNameInstAsst).Update(ctx, updateCond.ToMapStr(), updateInst)
			if err != nil {
				return err
			}

		} else {
			asst.AsstKindID = common.AssociationTypeDefault
			property := properyMap[buildObjPropertyMapKey(asst.ObjectID, asst.ObjectAttID)]
			switch property.PropertyType {
			case "singleasst":
				asst.Mapping = metadata.OneToManyMapping
			case "multiasst":
				asst.Mapping = metadata.ManyToManyMapping
			default:
				blog.Warnf("property: %+v, asst: %+v, for key: %v", property, asst,
					buildObjPropertyMapKey(asst.ObjectID, asst.ObjectAttID))
				asst.Mapping = metadata.ManyToManyMapping
			}
			// 交换 源<->目标
			asst.AssociationAliasName = property.PropertyName
			asst.ObjectID, asst.AsstObjID = asst.AsstObjID, asst.ObjectID
			asst.OnDelete = metadata.NoAction
			asst.IsPre = pfalse()
			asst.AssociationName = buildObjAsstID(asst)

			blog.Infof("obj: %s, att: %s to asst %+v", asst.ObjectID, asst.ObjectAttID, asst)
			// update ObjAsst
			updateCond := condition.CreateCondition()
			updateCond.Field("id").Eq(asst.ID)
			if err = db.Table(common.BKTableNameObjAsst).Update(ctx, updateCond.ToMapStr(), asst); err != nil {
				return err
			}

			instCond := condition.CreateCondition()
			instCond.Field("bk_obj_id").Eq(asst.AsstObjID)
			instCond.Field("bk_asst_obj_id").Eq(asst.ObjectID)
			instCond.Field(flag).NotEq(true)

			pageSize := uint64(2000)
			page := 0
			for {
				page += 1
				// update ObjAsst
				instAssts := []metadata.InstAsst{}
				blog.Infof("find data from table: %s, page: %d, cond: %+v", common.BKTableNameInstAsst, page,
					instCond.ToMapStr())
				if err = db.Table(common.BKTableNameInstAsst).Find(instCond.ToMapStr()).Limit(pageSize).All(ctx,
					&instAssts); err != nil {
					return err
				}

				blog.Infof("find data from table: %s, cond: %+v, result count: %d", common.BKTableNameInstAsst,
					instCond.ToMapStr(), len(instAssts))
				if len(instAssts) == 0 {
					break
				}
				for _, instAsst := range instAssts {
					updateInst := mapstr.New()
					updateInst.Set("bk_obj_asst_id", asst.AssociationName)
					updateInst.Set("bk_asst_id", asst.AsstKindID)

					// 交换 源<->目标
					updateInst.Set("bk_obj_id", instAsst.AsstObjectID)
					updateInst.Set("bk_asst_obj_id", instAsst.ObjectID)
					updateInst.Set("bk_inst_id", instAsst.AsstInstID)
					updateInst.Set("bk_asst_inst_id", instAsst.InstID)

					updateInst.Set(flag, true)

					updateInst.Set("last_time", time.Now())
					blog.Infof("update instasst, id: %d, updateInst: %+v", instAsst.ID, updateInst)
					if err = db.Table(common.BKTableNameInstAsst).Update(ctx,
						mapstr.MapStr{
							"id": instAsst.ID,
						}, updateInst); err != nil {
						return err
					}
				}
			}

		}
	}
	blog.Infof("start drop column cond: %s", flag)
	if err = dropFlagColumn(ctx, db, conf); err != nil {
		return err
	}

	// update bk_cloud_id to int
	cloudIDUpdateCond := condition.CreateCondition()
	cloudIDUpdateCond.Field(common.BKObjIDField).Eq(common.BKInnerObjIDHost)
	cloudIDUpdateCond.Field(common.BKPropertyIDField).Eq(common.BKCloudIDField)
	cloudIDUpdateData := mapstr.New()
	cloudIDUpdateData.Set(common.BKPropertyTypeField, "foreignkey")
	cloudIDUpdateData.Set(common.BKOptionField, nil)
	blog.Infof("update host cloud association cond: %+v, data: %+v", cloudIDUpdateCond.ToMapStr(), cloudIDUpdateData)
	err = db.Table(common.BKTableNameObjAttDes).Update(ctx, cloudIDUpdateCond.ToMapStr(), cloudIDUpdateData)
	if err != nil {
		return err
	}
	deleteHostCloudAssociation := condition.CreateCondition()
	deleteHostCloudAssociation.Field("bk_obj_id").Eq(common.BKInnerObjIDHost)
	deleteHostCloudAssociation.Field("bk_asst_obj_id").Eq(common.BKInnerObjIDPlat)
	blog.Infof("delete host cloud association table: %s, cond: %+v", common.BKTableNameObjAsst,
		deleteHostCloudAssociation.ToMapStr())
	err = db.Table(common.BKTableNameObjAsst).Delete(ctx, deleteHostCloudAssociation.ToMapStr())
	if err != nil {
		return err
	}

	blog.Infof("delete host cloud association table: %s, cond: %v", common.BKTableNameObjAttDes,
		propertyCond.ToMapStr())
	// drop outdate propertys
	err = db.Table(common.BKTableNameObjAttDes).Delete(ctx, propertyCond.ToMapStr())
	if err != nil {
		return err
	}

	// drop outdate column
	outdateColumns := []string{"bk_object_att_id", "bk_asst_forward", "bk_asst_name"}
	for _, column := range outdateColumns {
		blog.Infof("delete field from table: %s, cond: %+v", common.BKTableNameObjAsst, column)
		if err = db.Table(common.BKTableNameObjAsst).DropColumn(ctx, column); err != nil {
			return err
		}
	}

	delCond := condition.CreateCondition()
	delCond.Field(common.AssociationKindIDField).Eq(nil)
	blog.Infof("delete host cloud association table: %s, cond: %+v", common.BKTableNameObjAsst, delCond.ToMapStr())
	if err = db.Table(common.BKTableNameObjAsst).Delete(ctx, delCond.ToMapStr()); err != nil {
		return err
	}
	return nil
}

func dropFlagColumn(ctx context.Context, db dal.RDB, conf *history.Config) error {
	flag := "updateflag"
	flagFilter := map[string]interface{}{
		flag: map[string]interface{}{
			"$exists": true,
		},
	}
	cnt, err := db.Table(common.BKTableNameInstAsst).Find(flagFilter).Count(ctx)
	if err != nil {
		blog.Errorf("dropFlagColumn failed, Find err: %v, filter: %#v, ", err, flagFilter)
		return err
	}

	if cnt == 0 {
		return nil
	}

	pageSize := uint64(2000)
	for startIdx := uint64(0); startIdx < cnt; startIdx += pageSize {
		insts := make([]map[string]int64, 0)
		if err := db.Table(common.BKTableNameInstAsst).Find(flagFilter).Fields(common.BKFieldID).Start(startIdx).Limit(pageSize).All(ctx,
			&insts); err != nil {
			blog.Errorf("find insts failed, Find err: %v", err)
			return err
		}
		instIDs := make([]int64, len(insts))
		for i, inst := range insts {
			instIDs[i] = inst[common.BKFieldID]
		}

		filter := map[string]interface{}{
			common.BKFieldID: map[string]interface{}{
				"$in": instIDs,
			},
		}
		if err := db.Table(common.BKTableNameInstAsst).DropDocsColumn(ctx, flag, filter); err != nil {
			blog.Errorf("dropFlagColumn failed, filter: %#v, err: %v", filter, err)
			return err
		}
	}
	blog.Infof("drop flag count: %d successfully", cnt)

	return nil
}

func buildObjAsstID(asst Association) string {
	return fmt.Sprintf("%s_%s_%s", asst.ObjectID, asst.AsstKindID, asst.AsstObjID)
}

func ptrue() *bool {
	tmp := true
	return &tmp
}
func pfalse() *bool {
	tmp := false
	return &tmp
}

// AssociationKind is the kind of the association, which is used to describe the relationship between two objects.
type AssociationKind struct {
	ID int64 `field:"id" json:"id" bson:"id"`
	// a unique association id created by user.
	AssociationKindID string `field:"bk_asst_id" json:"bk_asst_id" bson:"bk_asst_id"`
	// a memorable name for this association kind, could be a chinese name, a english name etc.
	AssociationKindName string `field:"bk_asst_name" json:"bk_asst_name" bson:"bk_asst_name"`
	// the owner that this association type belongs to.
	OwnerID string `field:"bk_supplier_account" json:"bk_supplier_account" bson:"bk_supplier_account"`
	// the describe for the relationship from source object to the target(destination) object, which will be displayed
	// when the topology is constructed between objects.
	SourceToDestinationNote string `field:"src_des" json:"src_des" bson:"src_des"`
	// the describe for the relationship from the target(destination) object to source object, which will be displayed
	// when the topology is constructed between objects.
	DestinationToSourceNote string `field:"dest_des" json:"dest_des" bson:"dest_des"`
	// the association direction between two objects.
	Direction metadata.AssociationDirection `field:"direction" json:"direction" bson:"direction"`
	// whether this is a pre-defined kind.
	IsPre *bool `field:"ispre" json:"ispre" bson:"ispre"`
}
