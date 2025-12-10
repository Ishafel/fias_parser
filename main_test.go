package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestDetectXSDRoot(t *testing.T) {
	root, err := detectXSDRoot(filepath.Join("gar_schemas", "AS_HOUSES_2_251_08_04_01_01.xsd"))
	if err != nil {
		t.Fatalf("detectXSDRoot returned error: %v", err)
	}
	if root != "HOUSES" {
		t.Fatalf("expected root HOUSES, got %s", root)
	}
}

func TestDetectXMLRoot(t *testing.T) {
	root, err := detectXMLRoot(filepath.Join("testdata", "houses.xml"))
	if err != nil {
		t.Fatalf("detectXMLRoot returned error: %v", err)
	}
	if root != "HOUSES" {
		t.Fatalf("expected root HOUSES, got %s", root)
	}
}

func TestStreamElementsDefaultTarget(t *testing.T) {
	var buf bytes.Buffer
	if err := streamElements(filepath.Join("testdata", "houses.xml"), "", &buf); err != nil {
		t.Fatalf("streamElements returned error: %v", err)
	}

	dec := json.NewDecoder(&buf)
	var records []record
	for dec.More() {
		var rec record
		if err := dec.Decode(&rec); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		records = append(records, rec)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].Element != "HOUSE" || records[0].Attributes["ID"] != "1" || records[0].Content != "Main" {
		t.Fatalf("unexpected first record: %+v", records[0])
	}
}

func TestStreamElementsWithExplicitElement(t *testing.T) {
	var buf bytes.Buffer
	if err := streamElements(filepath.Join("testdata", "address_objects.xml"), "OBJECT", &buf); err != nil {
		t.Fatalf("streamElements returned error: %v", err)
	}

	dec := json.NewDecoder(&buf)
	count := 0
	for dec.More() {
		var rec record
		if err := dec.Decode(&rec); err != nil {
			t.Fatalf("decode json: %v", err)
		}
		if rec.Element != "OBJECT" {
			t.Fatalf("expected element OBJECT, got %s", rec.Element)
		}
		count++
	}

	if count != 2 {
		t.Fatalf("expected 2 OBJECT elements, got %d", count)
	}
}
