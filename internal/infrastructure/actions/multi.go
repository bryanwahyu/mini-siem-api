package actions

import "context"

type Multi struct{ Ports []interface{ ApplyBlock(context.Context,string,string) error; ApplyUnblock(context.Context,string,string) error } }

func (m *Multi) ApplyBlock(ctx context.Context, ip, reason string) error {
    var err error
    for _, p := range m.Ports { if e := p.ApplyBlock(ctx, ip, reason); e != nil { err = e } }
    return err
}

func (m *Multi) ApplyUnblock(ctx context.Context, ip, reason string) error {
    var err error
    for _, p := range m.Ports { if e := p.ApplyUnblock(ctx, ip, reason); e != nil { err = e } }
    return err
}

