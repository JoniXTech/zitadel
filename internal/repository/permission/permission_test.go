package permission

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zitadel/zitadel/internal/eventstore"
)

func TestRemovedEvent_Fields(t *testing.T) {
	agg := NewAggregate("instance1")
	ops := NewRemovedEvent(context.Background(), agg, "IAM_OWNER", "group.read").Fields()

	assert.Equal(t, []*eventstore.FieldOperation{{
		Remove: map[eventstore.FieldType]any{
			eventstore.FieldTypeInstanceID:     "instance1",
			eventstore.FieldTypeResourceOwner:  "instance1",
			eventstore.FieldTypeAggregateType:  AggregateType,
			eventstore.FieldTypeAggregateID:    "instance1",
			eventstore.FieldTypeObjectType:     RolePermissionType,
			eventstore.FieldTypeObjectID:       "IAM_OWNER",
			eventstore.FieldTypeObjectRevision: RolePermissionRevision,
			eventstore.FieldTypeFieldName:      PermissionSearchField,
			eventstore.FieldTypeValue:          "group.read",
		},
	}}, ops)
}
