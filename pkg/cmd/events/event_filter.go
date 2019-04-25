package events

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

type EventFilter interface {
	FilterEvents(events ...*corev1.Event) []*corev1.Event
}

type EventFilters []EventFilter

func (f EventFilters) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := make([]*corev1.Event, len(events), len(events))
	copy(ret, events)

	for _, filter := range f {
		ret = filter.FilterEvents(ret...)
	}

	return ret
}

type FilterByWarnings struct {
}

func (f *FilterByWarnings) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := []*corev1.Event{}
	for i := range events {
		event := events[i]
		if event.Type != corev1.EventTypeWarning {
			continue
		}
		ret = append(ret, event)
	}

	return ret
}

type FilterByNamespaces struct {
	Namespaces sets.String
}

func (f *FilterByNamespaces) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := []*corev1.Event{}
	for i := range events {
		event := events[i]
		ns := event.InvolvedObject.Namespace
		// check for an anti-match
		if f.Namespaces.Has("-" + ns) {
			continue
		}
		if f.Namespaces.Has(ns) {
			ret = append(ret, event)
		}
	}

	return ret
}

type FilterByNames struct {
	Names sets.String
}

func (f *FilterByNames) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := []*corev1.Event{}
	for i := range events {
		event := events[i]
		name := event.InvolvedObject.Name
		// check for an anti-match
		if f.Names.Has("-" + name) {
			continue
		}
		if f.Names.Has(name) {
			ret = append(ret, event)
		}
	}

	return ret
}

type FilterByReasons struct {
	Reasons sets.String
}

func (f *FilterByReasons) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := []*corev1.Event{}
	for i := range events {
		event := events[i]
		reason := event.Reason
		// check for an anti-match
		if f.Reasons.Has("-" + reason) {
			continue
		}
		if f.Reasons.Has(reason) {
			ret = append(ret, event)
		}
	}

	return ret
}

type FilterByUIDs struct {
	UIDs sets.String
}

func (f *FilterByUIDs) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := []*corev1.Event{}
	for i := range events {
		event := events[i]
		currUID := string(event.InvolvedObject.UID)
		// check for an anti-match
		if f.UIDs.Has("-" + currUID) {
			continue
		}
		if f.UIDs.Has(currUID) {
			ret = append(ret, event)
		}
	}

	return ret
}

type FilterByComponent struct {
	Components sets.String
}

func (f *FilterByComponent) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := []*corev1.Event{}
	for i := range events {
		event := events[i]
		// check for an anti-match
		if f.Components.Has("-" + event.ReportingController) {
			continue
		}
		if f.Components.Has(event.ReportingController) {
			ret = append(ret, event)
		}
	}

	return ret
}

type FilterByKind struct {
	Kinds map[schema.GroupKind]bool
}

func (f *FilterByKind) FilterEvents(events ...*corev1.Event) []*corev1.Event {
	ret := []*corev1.Event{}
	for i := range events {
		event := events[i]
		gv, _ := schema.ParseGroupVersion(event.InvolvedObject.APIVersion)
		gk := gv.WithKind(event.InvolvedObject.Kind).GroupKind()
		antiMatch := schema.GroupKind{Kind: "-" + gk.Kind, Group: gk.Group}

		// check for an anti-match
		if f.Kinds[antiMatch] {
			continue
		}
		if f.Kinds[gk] {
			ret = append(ret, event)
		}

		// if we aren't an exact match, match on resource only if group is '*'
		// check for an anti-match
		antiMatched := false
		for currKind := range f.Kinds {
			if currKind.Group == "*" && currKind.Kind == antiMatch.Kind {
				antiMatched = true
				break
			}
			if currKind.Kind == "-*" && currKind.Group == gk.Group {
				antiMatched = true
				break
			}
		}
		if antiMatched {
			continue
		}

		for currResource := range f.Kinds {
			if currResource.Group == "*" && currResource.Kind == "*" {
				ret = append(ret, event)
				break
			}
			if currResource.Group == "*" && currResource.Kind == gk.Kind {
				ret = append(ret, event)
				break
			}
			if currResource.Kind == "*" && currResource.Group == gk.Group {
				ret = append(ret, event)
				break
			}
		}
	}

	return ret
}