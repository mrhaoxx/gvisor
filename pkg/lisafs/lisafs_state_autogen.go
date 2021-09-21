// automatically generated by stateify.

package lisafs

import (
	"gvisor.dev/gvisor/pkg/state"
)

func (l *controlFDList) StateTypeName() string {
	return "pkg/lisafs.controlFDList"
}

func (l *controlFDList) StateFields() []string {
	return []string{
		"head",
		"tail",
	}
}

func (l *controlFDList) beforeSave() {}

// +checklocksignore
func (l *controlFDList) StateSave(stateSinkObject state.Sink) {
	l.beforeSave()
	stateSinkObject.Save(0, &l.head)
	stateSinkObject.Save(1, &l.tail)
}

func (l *controlFDList) afterLoad() {}

// +checklocksignore
func (l *controlFDList) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &l.head)
	stateSourceObject.Load(1, &l.tail)
}

func (e *controlFDEntry) StateTypeName() string {
	return "pkg/lisafs.controlFDEntry"
}

func (e *controlFDEntry) StateFields() []string {
	return []string{
		"next",
		"prev",
	}
}

func (e *controlFDEntry) beforeSave() {}

// +checklocksignore
func (e *controlFDEntry) StateSave(stateSinkObject state.Sink) {
	e.beforeSave()
	stateSinkObject.Save(0, &e.next)
	stateSinkObject.Save(1, &e.prev)
}

func (e *controlFDEntry) afterLoad() {}

// +checklocksignore
func (e *controlFDEntry) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &e.next)
	stateSourceObject.Load(1, &e.prev)
}

func (r *controlFDRefs) StateTypeName() string {
	return "pkg/lisafs.controlFDRefs"
}

func (r *controlFDRefs) StateFields() []string {
	return []string{
		"refCount",
	}
}

func (r *controlFDRefs) beforeSave() {}

// +checklocksignore
func (r *controlFDRefs) StateSave(stateSinkObject state.Sink) {
	r.beforeSave()
	stateSinkObject.Save(0, &r.refCount)
}

// +checklocksignore
func (r *controlFDRefs) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &r.refCount)
	stateSourceObject.AfterLoad(r.afterLoad)
}

func (l *openFDList) StateTypeName() string {
	return "pkg/lisafs.openFDList"
}

func (l *openFDList) StateFields() []string {
	return []string{
		"head",
		"tail",
	}
}

func (l *openFDList) beforeSave() {}

// +checklocksignore
func (l *openFDList) StateSave(stateSinkObject state.Sink) {
	l.beforeSave()
	stateSinkObject.Save(0, &l.head)
	stateSinkObject.Save(1, &l.tail)
}

func (l *openFDList) afterLoad() {}

// +checklocksignore
func (l *openFDList) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &l.head)
	stateSourceObject.Load(1, &l.tail)
}

func (e *openFDEntry) StateTypeName() string {
	return "pkg/lisafs.openFDEntry"
}

func (e *openFDEntry) StateFields() []string {
	return []string{
		"next",
		"prev",
	}
}

func (e *openFDEntry) beforeSave() {}

// +checklocksignore
func (e *openFDEntry) StateSave(stateSinkObject state.Sink) {
	e.beforeSave()
	stateSinkObject.Save(0, &e.next)
	stateSinkObject.Save(1, &e.prev)
}

func (e *openFDEntry) afterLoad() {}

// +checklocksignore
func (e *openFDEntry) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &e.next)
	stateSourceObject.Load(1, &e.prev)
}

func (r *openFDRefs) StateTypeName() string {
	return "pkg/lisafs.openFDRefs"
}

func (r *openFDRefs) StateFields() []string {
	return []string{
		"refCount",
	}
}

func (r *openFDRefs) beforeSave() {}

// +checklocksignore
func (r *openFDRefs) StateSave(stateSinkObject state.Sink) {
	r.beforeSave()
	stateSinkObject.Save(0, &r.refCount)
}

// +checklocksignore
func (r *openFDRefs) StateLoad(stateSourceObject state.Source) {
	stateSourceObject.Load(0, &r.refCount)
	stateSourceObject.AfterLoad(r.afterLoad)
}

func init() {
	state.Register((*controlFDList)(nil))
	state.Register((*controlFDEntry)(nil))
	state.Register((*controlFDRefs)(nil))
	state.Register((*openFDList)(nil))
	state.Register((*openFDEntry)(nil))
	state.Register((*openFDRefs)(nil))
}
