package lfs

import "github.com/git-lfs/git-lfs/api"

type Downloadable struct {
	pointer *WrappedPointer
	object  *api.ObjectResource
}

func (d *Downloadable) Object() *api.ObjectResource {
	return d.object
}

func (d *Downloadable) Oid() string {
	return d.pointer.Oid
}

func (d *Downloadable) Size() int64 {
	return d.pointer.Size
}

func (d *Downloadable) Name() string {
	return d.pointer.Name
}

func (d *Downloadable) Path() string {
	p, _ := LocalMediaPath(d.pointer.Oid)
	return p
}

func (d *Downloadable) SetObject(o *api.ObjectResource) {
	d.object = o
}

func NewDownloadable(p *WrappedPointer) *Downloadable {
	return &Downloadable{pointer: p}
}
