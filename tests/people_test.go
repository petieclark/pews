package tests

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPeopleCreate(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	t.Run("create person successfully", func(t *testing.T) {
		resp, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john.doe@example.com",
			"phone":      "555-0100",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected id in response")
		}
		assertEqual(t, "John", result["first_name"])
		assertEqual(t, "Doe", result["last_name"])
		assertEqual(t, "john.doe@example.com", result["email"])
	})

	t.Run("create person with minimal fields", func(t *testing.T) {
		resp, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "Jane",
			"last_name":  "Smith",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, "Jane", result["first_name"])
		assertEqual(t, "Smith", result["last_name"])
	})

	t.Run("create person with all fields", func(t *testing.T) {
		resp, err := ts.post("/api/people", map[string]interface{}{
			"first_name":        "Bob",
			"last_name":         "Johnson",
			"email":             "bob@example.com",
			"phone":             "555-0200",
			"address_line1":     "123 Main St",
			"city":              "Springfield",
			"state":             "IL",
			"zip":               "62701",
			"birthdate":         "1980-05-15",
			"gender":            "male",
			"membership_status": "member",
			"notes":             "Regular attender",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, "Bob", result["first_name"])
		assertEqual(t, "123 Main St", result["address_line1"])
		assertEqual(t, "member", result["membership_status"])
	})
}

func TestPeopleList(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create test people
	people := []map[string]interface{}{
		{"first_name": "Alice", "last_name": "Anderson"},
		{"first_name": "Bob", "last_name": "Brown"},
		{"first_name": "Charlie", "last_name": "Clark"},
		{"first_name": "Diana", "last_name": "Davis"},
		{"first_name": "Eve", "last_name": "Evans"},
	}

	for _, person := range people {
		resp, err := ts.post("/api/people", person)
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)
	}

	t.Run("list all people", func(t *testing.T) {
		resp, err := ts.get("/api/people")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		peopleList := result["people"].([]interface{})
		if len(peopleList) != 5 {
			t.Fatalf("Expected 5 people, got %d", len(peopleList))
		}

		total := result["total"].(float64)
		if total != 5 {
			t.Fatalf("Expected total 5, got %v", total)
		}
	})

	t.Run("list with pagination", func(t *testing.T) {
		resp, err := ts.get("/api/people?page=1&limit=2")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		peopleList := result["people"].([]interface{})
		if len(peopleList) != 2 {
			t.Fatalf("Expected 2 people, got %d", len(peopleList))
		}

		total := result["total"].(float64)
		if total != 5 {
			t.Fatalf("Expected total 5, got %v", total)
		}
	})

	t.Run("search by name", func(t *testing.T) {
		resp, err := ts.get("/api/people?q=bob")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		peopleList := result["people"].([]interface{})
		if len(peopleList) == 0 {
			t.Fatal("Expected to find Bob in search results")
		}

		person := peopleList[0].(map[string]interface{})
		if person["first_name"] != "Bob" {
			t.Fatalf("Expected Bob, got %v", person["first_name"])
		}
	})
}

func TestPeopleGet(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a person
	resp, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john@example.com",
	})
	assertNoError(t, err)
	assertStatusCode(t, resp, http.StatusCreated)

	var createResult map[string]interface{}
	err = decodeJSON(resp, &createResult)
	assertNoError(t, err)
	personID := createResult["id"].(string)

	t.Run("get person by id", func(t *testing.T) {
		resp, err := ts.get(fmt.Sprintf("/api/people/%s", personID))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, personID, result["id"])
		assertEqual(t, "John", result["first_name"])
		assertEqual(t, "Doe", result["last_name"])
	})

	t.Run("get nonexistent person", func(t *testing.T) {
		resp, err := ts.get("/api/people/00000000-0000-0000-0000-000000000000")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusNotFound)
	})
}

func TestPeopleUpdate(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a person
	resp, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john@example.com",
	})
	assertNoError(t, err)
	assertStatusCode(t, resp, http.StatusCreated)

	var createResult map[string]interface{}
	err = decodeJSON(resp, &createResult)
	assertNoError(t, err)
	personID := createResult["id"].(string)

	t.Run("update person", func(t *testing.T) {
		resp, err := ts.put(fmt.Sprintf("/api/people/%s", personID), map[string]interface{}{
			"first_name": "Jane",
			"last_name":  "Smith",
			"email":      "jane.smith@example.com",
			"phone":      "555-0100",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, "Jane", result["first_name"])
		assertEqual(t, "Smith", result["last_name"])
		assertEqual(t, "jane.smith@example.com", result["email"])
		assertEqual(t, "555-0100", result["phone"])

		// Verify the update persisted
		getResp, err := ts.get(fmt.Sprintf("/api/people/%s", personID))
		assertNoError(t, err)
		assertStatusCode(t, getResp, http.StatusOK)

		var getResult map[string]interface{}
		err = decodeJSON(getResp, &getResult)
		assertNoError(t, err)

		assertEqual(t, "Jane", getResult["first_name"])
	})
}

func TestPeopleDelete(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a person
	resp, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
	})
	assertNoError(t, err)
	assertStatusCode(t, resp, http.StatusCreated)

	var createResult map[string]interface{}
	err = decodeJSON(resp, &createResult)
	assertNoError(t, err)
	personID := createResult["id"].(string)

	t.Run("delete person", func(t *testing.T) {
		resp, err := ts.delete(fmt.Sprintf("/api/people/%s", personID))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		// Verify person is deleted
		getResp, err := ts.get(fmt.Sprintf("/api/people/%s", personID))
		assertNoError(t, err)
		assertStatusCode(t, getResp, http.StatusNotFound)
	})

	t.Run("delete nonexistent person", func(t *testing.T) {
		resp, err := ts.delete("/api/people/00000000-0000-0000-0000-000000000000")
		assertNoError(t, err)
		// Should not error, just return 404 or 200
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 200 or 404, got %d", resp.StatusCode)
		}
	})
}
