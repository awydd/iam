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
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "user",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("user"),
		entsql.WithComments(true),
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			MaxLen(32).
			NotEmpty().
			Unique(),

		field.Bytes("password").
			NotEmpty().
			Sensitive().
			MaxLen(255).
			SchemaType(map[string]string{
				"mysql": "varbinary(255)",
			}),

		field.UUID("uuid", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Unique(),

		field.String("email").
			MinLen(6).
			MaxLen(128).
			NotEmpty().
			Unique(),

		field.Uint8("status").
			GoType(enum.UserStatus(0)).
			Default(uint8(enum.UserStatusPending)),

		field.Time("last_login_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// 删除用户时清理令牌
		edge.To("tokens", Token.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("last_login_at"),
	}
}
