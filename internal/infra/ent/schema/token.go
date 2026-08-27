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

// Token holds the schema definition for the Token entity.
type Token struct {
	ent.Schema
}

func (Token) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "token",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("token"),
		entsql.WithComments(true),
	}
}

// Fields of the Token.
func (Token) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),

		field.Int("application_id"),

		field.Uint8("type").
			GoType(enum.TokenType(0)).
			Default(uint8(enum.TokenTypeRefresh)),

		field.UUID("jti", uuid.UUID{}).
			Immutable().
			Optional().
			Unique(),

		field.UUID("session_id", uuid.UUID{}),

		field.Bytes("identity").
			Unique().
			MaxLen(32).
			SchemaType(map[string]string{
				"mysql": "binary(32)",
			}),

		field.Time("expires_at").
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Time("revoked_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Time("last_active_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.String("ip").
			MaxLen(45).
			Optional(),

		field.Text("user_agent").
			MaxLen(512).
			Optional(),
	}
}

func (Token) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
	}
}

// Edges of the Token.
func (Token) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("tokens").
			Field("user_id").
			Required().
			Unique(),

		edge.From("application", Application.Type).
			Ref("tokens").
			Field("application_id").
			Required().
			Unique(),
	}
}

func (Token) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("user_id", "type"),
		index.Fields("user_id", "application_id"),
		index.Fields("user_id", "session_id"),
		index.Fields("session_id"),
		index.Fields("expires_at"),
	}
}
