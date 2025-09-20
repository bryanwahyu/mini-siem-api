package notify

import "context"

type Multi struct{ Notifiers []interface{ Notify(context.Context,string,string) error } }

func (m *Multi) Notify(ctx context.Context, title, text string) error {
    var err error
    for _, n := range m.Notifiers { if e := n.Notify(ctx, title, text); e != nil { err = e } }
    return err
}

