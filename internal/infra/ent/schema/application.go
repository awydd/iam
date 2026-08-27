package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/awydd/iam/internal/enum"
	"github.com/awydd/iam/internal/infra/ent/mixin"
)

// Application holds the schema definition for the Application entity.
type Application struct {
	ent.Schema
}

func (Application) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "application",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("application"),
		entsql.WithComments(true),
	}
}

// Fields of the Application.
func (Application) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			NotEmpty().
			MaxLen(64),

		field.String("client_id").
			Unique().
			MaxLen(36).
			NotEmpty(),

		field.Bytes("client_secret").
			Optional().
			MaxLen(32).
			Sensitive().
			SchemaType(map[string]string{
				"mysql": "binary(32)",
			}),

		field.JSON("redirect_uris", []string{}),

		field.Uint8("status").
			GoType(enum.ApplicationStatus(0)).
			Default(uint8(enum.ApplicationStatusActive)),

		field.Uint8("type").
			GoType(enum.ApplicationClientType(0)).
			Default(uint8(enum.ApplicationClientTypePublic)),

		field.Int("access_token_ttl").
			Min(60).
			Max(int(enum.ApplicationMaxAccessTokenTTL.Seconds())).
			Default(int(enum.ApplicationDefaultAccessTokenTTL.Seconds())),

		field.Int("refresh_token_ttl").
			Min(3600).
			Max(int(enum.ApplicationMaxRefreshTokenTTL.Seconds())).
			Default(int(enum.ApplicationDefaultRefreshTokenTTL.Seconds())),

		field.Bool("is_system").
			Immutable().
			Default(false),
	}
}

func (Application) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Edges of the Application.
func (Application) Edges() []ent.Edge {
	return []ent.Edge{
		// 删除应用时清除令牌
		edge.To("tokens", Token.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Application) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("type"),
		index.Fields("status", "type"),
	}
}
