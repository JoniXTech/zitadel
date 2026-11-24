package idp

import (
	"github.com/zitadel/zitadel/internal/crypto"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/zerrors"
)

type DiscordIDPAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	ID           string              `json:"id"`
	Name         string              `json:"name,omitempty"`
	ClientID     string              `json:"clientId"`
	ClientSecret *crypto.CryptoValue `json:"clientSecret"`
	Scopes       []string            `json:"scopes,omitempty"`
	Prompt       string              `json:"prompt,omitempty"`
	Options
}

func NewDiscordIDPAddedEvent(
	base *eventstore.BaseEvent,
	id,
	name,
	clientID string,
	clientSecret *crypto.CryptoValue,
	scopes []string,
	prompt string,
	options Options,
) *DiscordIDPAddedEvent {
	return &DiscordIDPAddedEvent{
		BaseEvent:    *base,
		ID:           id,
		Name:         name,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Prompt:       prompt,
		Options:      options,
	}
}

func (e *DiscordIDPAddedEvent) Payload() interface{} {
	return e
}

func (e *DiscordIDPAddedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return nil
}

func DiscordIDPAddedEventMapper(event eventstore.Event) (eventstore.Event, error) {
	e := &DiscordIDPAddedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := event.Unmarshal(e)
	if err != nil {
		return nil, zerrors.ThrowInternal(err, "IDP-SAff1", "unable to unmarshal event")
	}

	return e, nil
}

type DiscordIDPChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	ID           string              `json:"id"`
	Name         *string             `json:"name,omitempty"`
	ClientID     *string             `json:"clientId,omitempty"`
	ClientSecret *crypto.CryptoValue `json:"clientSecret,omitempty"`
	Scopes       []string            `json:"scopes,omitempty"`
	Prompt       *string             `json:"prompt,omitempty"`
	OptionChanges
}

func NewDiscordIDPChangedEvent(
	base *eventstore.BaseEvent,
	id string,
	changes []DiscordIDPChanges,
) (*DiscordIDPChangedEvent, error) {
	if len(changes) == 0 {
		return nil, zerrors.ThrowPreconditionFailed(nil, "IDP-Dg3qs", "zerrors.NoChangesFound")
	}
	changedEvent := &DiscordIDPChangedEvent{
		BaseEvent: *base,
		ID:        id,
	}
	for _, change := range changes {
		change(changedEvent)
	}
	return changedEvent, nil
}

type DiscordIDPChanges func(*DiscordIDPChangedEvent)

func ChangeDiscordName(name string) func(*DiscordIDPChangedEvent) {
	return func(e *DiscordIDPChangedEvent) {
		e.Name = &name
	}
}
func ChangeDiscordClientID(clientID string) func(*DiscordIDPChangedEvent) {
	return func(e *DiscordIDPChangedEvent) {
		e.ClientID = &clientID
	}
}

func ChangeDiscordClientSecret(clientSecret *crypto.CryptoValue) func(*DiscordIDPChangedEvent) {
	return func(e *DiscordIDPChangedEvent) {
		e.ClientSecret = clientSecret
	}
}

func ChangeDiscordScopes(scopes []string) func(*DiscordIDPChangedEvent) {
	return func(e *DiscordIDPChangedEvent) {
		e.Scopes = scopes
	}
}

func ChangeDiscordPrompt(prompt string) func(*DiscordIDPChangedEvent) {
	return func(e *DiscordIDPChangedEvent) {
		e.Prompt = &prompt
	}
}

func ChangeDiscordOptions(options OptionChanges) func(*DiscordIDPChangedEvent) {
	return func(e *DiscordIDPChangedEvent) {
		e.OptionChanges = options
	}
}

func (e *DiscordIDPChangedEvent) Payload() interface{} {
	return e
}

func (e *DiscordIDPChangedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return nil
}

func DiscordIDPChangedEventMapper(event eventstore.Event) (eventstore.Event, error) {
	e := &DiscordIDPChangedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := event.Unmarshal(e)
	if err != nil {
		return nil, zerrors.ThrowInternal(err, "IDP-SF3t2", "unable to unmarshal event")
	}

	return e, nil
}
