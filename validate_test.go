package ics

import (
	"strings"
	"testing"
)

func TestCalendarValidate_Empty(t *testing.T) {
	cal := NewCalendar()
	if err := cal.Validate(); err != nil {
		t.Fatalf("empty calendar should validate, got: %v", err)
	}
}

func TestCalendarValidate_VEventOnlyUID(t *testing.T) {
	cal := NewCalendar()
	cal.Components = append(cal.Components, NewEvent("evt-only-uid"))
	err := cal.Validate()
	if err == nil {
		t.Fatal("expected error for VEVENT missing DTSTAMP, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "VEVENT (uid=evt-only-uid)") {
		t.Errorf("error should identify component and uid, got: %q", msg)
	}
	if !strings.Contains(msg, "DTSTAMP") {
		t.Errorf("error should mention DTSTAMP, got: %q", msg)
	}
}

// Pins current Required() behaviour: DTSTART is required when the VEvent
// itself has no METHOD (not the enclosing Calendar, as RFC 5545 §3.6.1 says).
func TestCalendarValidate_VEventUIDDtstamp(t *testing.T) {
	cal := NewCalendar()
	evt := NewEvent("evt-uid-dtstamp")
	evt.SetProperty(ComponentPropertyDtstamp, "20240101T000000Z")
	cal.Components = append(cal.Components, evt)
	err := cal.Validate()
	if err == nil {
		t.Fatal("expected error: Required() currently flags DTSTART when the VEVENT has no METHOD")
	}
	if !strings.Contains(err.Error(), "DTSTART") {
		t.Errorf("error should mention DTSTART, got: %q", err.Error())
	}
}

func TestCalendarValidate_VEventUIDDtstampMethod(t *testing.T) {
	cal := NewCalendar()
	evt := NewEvent("evt-with-method")
	evt.SetProperty(ComponentPropertyDtstamp, "20240101T000000Z")
	evt.SetProperty(ComponentPropertyMethod, "PUBLISH")
	cal.Components = append(cal.Components, evt)
	if err := cal.Validate(); err != nil {
		t.Fatalf("event with METHOD should not require DTSTART per current Required(), got: %v", err)
	}
}

func TestCalendarValidate_VEventUIDDtstampDtstart(t *testing.T) {
	cal := NewCalendar()
	evt := NewEvent("evt-full")
	evt.SetProperty(ComponentPropertyDtstamp, "20240101T000000Z")
	evt.SetProperty(ComponentPropertyDtStart, "20240101T010000Z")
	cal.Components = append(cal.Components, evt)
	if err := cal.Validate(); err != nil {
		t.Fatalf("event with UID, DTSTAMP, DTSTART should validate, got: %v", err)
	}
}

// TestCalendarValidate_AggregatesAllMisses pins the errors.Join behaviour:
// exactly one line per missing property, joined by newlines.
func TestCalendarValidate_AggregatesAllMisses(t *testing.T) {
	cal := NewCalendar()
	cal.Components = append(cal.Components, &VEvent{})
	err := cal.Validate()
	if err == nil {
		t.Fatal("expected error for VEVENT missing UID, DTSTAMP, and DTSTART")
	}
	msg := err.Error()
	lines := strings.Split(msg, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected exactly 3 joined error lines, got %d: %q", len(lines), msg)
	}
	for _, want := range []string{"UID", "DTSTAMP", "DTSTART"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error should mention %s; got: %q", want, msg)
		}
	}
}

func TestCalendarValidate_NilAndTypedNilComponents(t *testing.T) {
	cal := NewCalendar()
	var typedNil *VEvent
	cal.Components = append(cal.Components, nil, typedNil)
	if err := cal.Validate(); err != nil {
		t.Fatalf("nil and typed-nil component entries should be skipped, got: %v", err)
	}
}

// TestCalendarValidate_RecursesIntoSubcomponents verifies that a VEVENT
// nested inside another VEVENT is also checked. Required() flags DTSTAMP on
// every *VEvent, so the inner one's miss must surface.
func TestCalendarValidate_RecursesIntoSubcomponents(t *testing.T) {
	cal := NewCalendar()
	outer := NewEvent("outer")
	outer.SetProperty(ComponentPropertyDtstamp, "20240101T000000Z")
	outer.SetProperty(ComponentPropertyDtStart, "20240101T010000Z")
	inner := NewEvent("inner-missing-dtstamp")
	inner.SetProperty(ComponentPropertyMethod, "PUBLISH") // suppress DTSTART requirement
	outer.Components = append(outer.Components, inner)
	cal.Components = append(cal.Components, outer)

	err := cal.Validate()
	if err == nil {
		t.Fatal("expected error from inner VEVENT missing DTSTAMP")
	}
	msg := err.Error()
	if !strings.Contains(msg, "uid=inner-missing-dtstamp") {
		t.Errorf("error should identify the inner component by uid, got: %q", msg)
	}
	if !strings.Contains(msg, "DTSTAMP") {
		t.Errorf("error should mention DTSTAMP, got: %q", msg)
	}
}

// TestCalendarValidate_GeneralComponent confirms that components the package
// doesn't model explicitly (parsed as *GeneralComponent) are walked rather
// than silently skipped. Required() returns false for them today, so the
// expectation is "no error, no panic."
func TestCalendarValidate_GeneralComponent(t *testing.T) {
	cal := NewCalendar()
	cal.Components = append(cal.Components, &GeneralComponent{Token: "X-VENDOR"})
	if err := cal.Validate(); err != nil {
		t.Fatalf("unknown component types should not produce errors today, got: %v", err)
	}
}
