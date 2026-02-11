package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCheckInStations(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	t.Run("create station", func(t *testing.T) {
		resp, err := ts.post("/api/checkins/stations", map[string]interface{}{
			"name":     "Main Lobby",
			"location": "Building A - First Floor",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected id in response")
		}
		assertEqual(t, "Main Lobby", result["name"])
		assertEqual(t, "Building A - First Floor", result["location"])
	})

	t.Run("list stations", func(t *testing.T) {
		// Create another station
		resp, err := ts.post("/api/checkins/stations", map[string]interface{}{
			"name":     "Children's Wing",
			"location": "Building B",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		// List all stations
		listResp, err := ts.get("/api/checkins/stations")
		assertNoError(t, err)
		assertStatusCode(t, listResp, http.StatusOK)

		var result []interface{}
		err = decodeJSON(listResp, &result)
		assertNoError(t, err)

		if len(result) < 2 {
			t.Fatalf("Expected at least 2 stations, got %d", len(result))
		}
	})

	t.Run("update station", func(t *testing.T) {
		// Create a station
		resp, err := ts.post("/api/checkins/stations", map[string]interface{}{
			"name":     "Youth Area",
			"location": "Building C",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var createResult map[string]interface{}
		err = decodeJSON(resp, &createResult)
		assertNoError(t, err)
		stationID := createResult["id"].(string)

		// Update the station
		updateResp, err := ts.put(fmt.Sprintf("/api/checkins/stations/%s", stationID), map[string]interface{}{
			"name":     "Youth & Young Adults",
			"location": "Building C - Updated",
		})
		assertNoError(t, err)
		assertStatusCode(t, updateResp, http.StatusOK)

		var result map[string]interface{}
		err = decodeJSON(updateResp, &result)
		assertNoError(t, err)

		assertEqual(t, "Youth & Young Adults", result["name"])
	})
}

func TestCheckInEvents(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	t.Run("create event", func(t *testing.T) {
		resp, err := ts.post("/api/checkins/events", map[string]interface{}{
			"name":       "Sunday Service",
			"start_time": time.Now().Format(time.RFC3339),
			"end_time":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected id in response")
		}
		assertEqual(t, "Sunday Service", result["name"])
	})

	t.Run("list events", func(t *testing.T) {
		// Create multiple events
		events := []map[string]interface{}{
			{
				"name":       "Wednesday Prayer",
				"start_time": time.Now().Format(time.RFC3339),
				"end_time":   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			},
			{
				"name":       "Youth Group",
				"start_time": time.Now().Format(time.RFC3339),
				"end_time":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
			},
		}

		for _, event := range events {
			resp, err := ts.post("/api/checkins/events", event)
			assertNoError(t, err)
			assertStatusCode(t, resp, http.StatusCreated)
		}

		// List all events
		resp, err := ts.get("/api/checkins/events")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if len(result) < 2 {
			t.Fatalf("Expected at least 2 events, got %d", len(result))
		}
	})

	t.Run("get event by id", func(t *testing.T) {
		// Create an event
		resp, err := ts.post("/api/checkins/events", map[string]interface{}{
			"name":       "Special Event",
			"start_time": time.Now().Format(time.RFC3339),
			"end_time":   time.Now().Add(3 * time.Hour).Format(time.RFC3339),
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var createResult map[string]interface{}
		err = decodeJSON(resp, &createResult)
		assertNoError(t, err)
		eventID := createResult["id"].(string)

		// Get the event
		getResp, err := ts.get(fmt.Sprintf("/api/checkins/events/%s", eventID))
		assertNoError(t, err)
		assertStatusCode(t, getResp, http.StatusOK)

		var result map[string]interface{}
		err = decodeJSON(getResp, &result)
		assertNoError(t, err)

		assertEqual(t, eventID, result["id"])
		assertEqual(t, "Special Event", result["name"])
	})

	t.Run("update event", func(t *testing.T) {
		// Create an event
		resp, err := ts.post("/api/checkins/events", map[string]interface{}{
			"name":       "Original Event",
			"start_time": time.Now().Format(time.RFC3339),
			"end_time":   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var createResult map[string]interface{}
		err = decodeJSON(resp, &createResult)
		assertNoError(t, err)
		eventID := createResult["id"].(string)

		// Update the event
		updateResp, err := ts.put(fmt.Sprintf("/api/checkins/events/%s", eventID), map[string]interface{}{
			"name":       "Updated Event Name",
			"start_time": time.Now().Format(time.RFC3339),
			"end_time":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		})
		assertNoError(t, err)
		assertStatusCode(t, updateResp, http.StatusOK)

		var result map[string]interface{}
		err = decodeJSON(updateResp, &result)
		assertNoError(t, err)

		assertEqual(t, "Updated Event Name", result["name"])
	})
}

func TestCheckInProcess(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a person
	personResp, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "John",
		"last_name":  "Attendee",
	})
	assertNoError(t, err)
	assertStatusCode(t, personResp, http.StatusOK)

	var personResult map[string]interface{}
	err = decodeJSON(personResp, &personResult)
	assertNoError(t, err)
	personID := personResult["id"].(string)

	// Create an event
	eventResp, err := ts.post("/api/checkins/events", map[string]interface{}{
		"name":       "Sunday Service",
		"start_time": time.Now().Format(time.RFC3339),
		"end_time":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	})
	assertNoError(t, err)
	assertStatusCode(t, eventResp, http.StatusOK)

	var eventResult map[string]interface{}
	err = decodeJSON(eventResp, &eventResult)
	assertNoError(t, err)
	eventID := eventResult["id"].(string)

	t.Run("check in person", func(t *testing.T) {
		resp, err := ts.post(fmt.Sprintf("/api/checkins/events/%s/checkin", eventID), map[string]interface{}{
			"person_id": personID,
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected check-in id in response")
		}
		assertEqual(t, personID, result["person_id"])
		assertEqual(t, eventID, result["event_id"])

		// Verify check_in_time is set
		if result["check_in_time"] == nil {
			t.Fatal("Expected check_in_time in response")
		}
	})

	t.Run("list attendees", func(t *testing.T) {
		// Check in another person
		person2Resp, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "Jane",
			"last_name":  "Attendee",
		})
		assertNoError(t, err)
		assertStatusCode(t, person2Resp, http.StatusOK)

		var person2Result map[string]interface{}
		err = decodeJSON(person2Resp, &person2Result)
		assertNoError(t, err)
		person2ID := person2Result["id"].(string)

		// Check in second person
		resp, err := ts.post(fmt.Sprintf("/api/checkins/events/%s/checkin", eventID), map[string]interface{}{
			"person_id": person2ID,
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		// List attendees
		listResp, err := ts.get(fmt.Sprintf("/api/checkins/events/%s/attendees", eventID))
		assertNoError(t, err)
		assertStatusCode(t, listResp, http.StatusOK)

		var result []interface{}
		err = decodeJSON(listResp, &result)
		assertNoError(t, err)

		if len(result) != 2 {
			t.Fatalf("Expected 2 attendees, got %d", len(result))
		}
	})

	t.Run("check out person", func(t *testing.T) {
		resp, err := ts.post(fmt.Sprintf("/api/checkins/events/%s/checkout", eventID), map[string]interface{}{
			"person_id": personID,
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		// Verify check_out_time is set
		if result["check_out_time"] == nil {
			t.Fatal("Expected check_out_time in response")
		}
	})

	t.Run("get person history", func(t *testing.T) {
		resp, err := ts.get(fmt.Sprintf("/api/checkins/person/%s/history", personID))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if len(result) == 0 {
			t.Fatal("Expected check-in history for person")
		}
	})
}

func TestCheckInStats(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create event and check in some people
	eventResp, err := ts.post("/api/checkins/events", map[string]interface{}{
		"name":       "Test Event",
		"start_time": time.Now().Format(time.RFC3339),
		"end_time":   time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	})
	assertNoError(t, err)
	assertStatusCode(t, eventResp, http.StatusOK)

	var eventResult map[string]interface{}
	err = decodeJSON(eventResp, &eventResult)
	assertNoError(t, err)
	eventID := eventResult["id"].(string)

	// Create and check in people
	for i := 0; i < 3; i++ {
		personResp, err := ts.post("/api/people", map[string]interface{}{
			"first_name": fmt.Sprintf("Person%d", i),
			"last_name":  "Test",
		})
		assertNoError(t, err)
		assertStatusCode(t, personResp, http.StatusOK)

		var personResult map[string]interface{}
		err = decodeJSON(personResp, &personResult)
		assertNoError(t, err)
		personID := personResult["id"].(string)

		// Check in
		checkinResp, err := ts.post(fmt.Sprintf("/api/checkins/events/%s/checkin", eventID), map[string]interface{}{
			"person_id": personID,
		})
		assertNoError(t, err)
		assertStatusCode(t, checkinResp, http.StatusOK)
	}

	t.Run("get check-in stats", func(t *testing.T) {
		resp, err := ts.get("/api/checkins/stats")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		// Stats should contain some data
		if result == nil {
			t.Fatal("Expected stats data in response")
		}
	})
}
