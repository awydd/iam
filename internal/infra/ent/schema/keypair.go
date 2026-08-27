package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/awydd/iam/internal/enum"
)

// Keypair holds the schema definition for the Keypair entity.
type Keypair struct {
	ent.Schema
}

func (Keypair) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "keypair",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		schema.Comment("keypair"),
		entsql.WithComments(true),
	}
}

// Fields of the Keypair.
func (Keypair) Fields() []ent.Field {
	return []ent.Field{
		field.String("kid").
			Unique().
			NotEmpty().
			MaxLen(64),

		field.Uint8("algorithm").
			GoType(enum.KeypairAlgorithm(0)).
			Default(uint8(enum.KeypairAlgoRS256)),

		field.Text("public_key").
			NotEmpty(),

		field.Text("private_key").
			NotEmpty().
			Sensitive(),

		field.Time("activated_at").
			Default(time.Now).
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Uint8("status").
			GoType(enum.KeypairStatus(0)).
			Default(uint8(enum.KeypairStatusActive)),

		field.Time("retire_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Keypair.
func (Keypair) Edges() []ent.Edge {
	return nil
}

func (Keypair) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("status", "activated_at"),
	}
}
