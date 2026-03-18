package auth

import "context"

type actorContextKey struct{}

// WithActor stores an Actor in the context.
func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

// ActorFromContext retrieves the Actor stored by WithActor. If no actor is
// present, a zero-value Actor with Type == ActorTypeNone is returned along
// with ok == false.
func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorContextKey{}).(Actor)
	if !ok {
		return Actor{Type: ActorTypeNone}, false
	}
	return actor, true
}
