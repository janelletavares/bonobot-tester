package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

const requestsCollectionName = "requests"

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)
		list := &models.Collection{
			Name: requestsCollectionName,
			Type: models.CollectionTypeBase,
			// any authenticated user
			CreateRule: types.Pointer("@request.auth.id != ''"),
			// only your own things
			ViewRule: types.Pointer("@request.auth.id = requester.id"),
			ListRule: types.Pointer("@request.auth.id = requester.id"),
			// only admins
			UpdateRule: nil,
			DeleteRule: nil,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "meeting_id",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Min: types.Pointer(1),
					},
				},
				&schema.SchemaField{
					Name:     "meeting_passcode",
					Type:     schema.FieldTypeText,
					Required: true,
					Options: &schema.TextOptions{
						Min: types.Pointer(1),
					},
				},
				&schema.SchemaField{
					Name:     "duration",
					Type:     schema.FieldTypeNumber,
					Required: true,
					Options: &schema.NumberOptions{
						NoDecimal: true,
						Min:       types.Pointer(float64(1)),
						Max:       types.Pointer(float64(60 * 45)), // free meeting length
					},
				},
				&schema.SchemaField{
					Name:     "count",
					Type:     schema.FieldTypeNumber,
					Required: true,
					Options: &schema.NumberOptions{
						Min: types.Pointer(float64(1)),
						Max: types.Pointer(float64(15)), // arbitrary... free on AWS?
					},
				},
				&schema.SchemaField{
					Name: "requester",
					Type: schema.FieldTypeRelation,
					Options: schema.RelationOptions{
						CollectionId: "users",
						MaxSelect:    types.Pointer(1),
					},
				},
			),
		}

		if err := dao.SaveCollection(list); err != nil {
			return err
		}
		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		list, err := dao.FindCollectionByNameOrId(requestsCollectionName)
		if err != nil {
			return err
		}
		if err := dao.DeleteCollection(list); err != nil {
			return err
		}
		return nil
	})
}
