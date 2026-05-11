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

// TestCalendarValidate_VEventUIDDtstamp documents what Required() does today,
// not what RFC 5545 §3.6.1 prescribes. Required() at calendar.go gates DTSTART
// on whether the *VEvent itself* carries a METHOD property, but RFC 5545 ties
// that condition to the enclosing Calendar's METHOD. Per the task brief we
// pin the test to the implementation and leave the fix to a separate change.
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

func TestComponentBaseValidate_DirectCall(t *testing.T) {
	evt := NewEvent("direct")
	err := evt.ComponentBase.Validate(evt)
	if err == nil {
		t.Fatal("expected error from direct ComponentBase.Validate, got nil")
	}
	if !strings.Contains(err.Error(), "DTSTAMP") {
		t.Errorf("error should mention DTSTAMP, got: %q", err.Error())
	}
}
