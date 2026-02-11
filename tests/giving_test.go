package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestGivingFunds(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	t.Run("create fund", func(t *testing.T) {
		resp, err := ts.post("/api/giving/funds", map[string]interface{}{
			"name":        "General Fund",
			"description": "General church operations",
			"is_default":  true,
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected id in response")
		}
		assertEqual(t, "General Fund", result["name"])
		assertEqual(t, "General church operations", result["description"])
		assertEqual(t, true, result["is_default"])
	})

	t.Run("list funds", func(t *testing.T) {
		// Create multiple funds
		funds := []map[string]interface{}{
			{"name": "Building Fund", "description": "Building renovations", "is_default": false},
			{"name": "Missions Fund", "description": "Missionary support", "is_default": false},
		}

		for _, fund := range funds {
			resp, err := ts.post("/api/giving/funds", fund)
			assertNoError(t, err)
			assertStatusCode(t, resp, http.StatusCreated)
		}

		// List all funds
		resp, err := ts.get("/api/giving/funds")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if len(result) < 2 {
			t.Fatalf("Expected at least 2 funds, got %d", len(result))
		}
	})

	t.Run("update fund", func(t *testing.T) {
		// Create a fund
		resp, err := ts.post("/api/giving/funds", map[string]interface{}{
			"name":        "Youth Fund",
			"description": "Youth activities",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var createResult map[string]interface{}
		err = decodeJSON(resp, &createResult)
		assertNoError(t, err)
		fundID := createResult["id"].(string)

		// Update the fund
		resp, err = ts.put(fmt.Sprintf("/api/giving/funds/%s", fundID), map[string]interface{}{
			"name":        "Youth & Children Fund",
			"description": "Youth and children's ministry",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, "Youth & Children Fund", result["name"])
		assertEqual(t, "Youth and children's ministry", result["description"])
	})
}

func TestGivingDonations(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a fund first
	fundResp, err := ts.post("/api/giving/funds", map[string]interface{}{
		"name":        "General Fund",
		"description": "General operations",
		"is_default":  true,
	})
	assertNoError(t, err)
	assertStatusCode(t, fundResp, http.StatusOK)

	var fundResult map[string]interface{}
	err = decodeJSON(fundResp, &fundResult)
	assertNoError(t, err)
	fundID := fundResult["id"].(string)

	// Create a person
	personResp, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "John",
		"last_name":  "Donor",
		"email":      "john@donor.com",
	})
	assertNoError(t, err)
	assertStatusCode(t, personResp, http.StatusOK)

	var personResult map[string]interface{}
	err = decodeJSON(personResp, &personResult)
	assertNoError(t, err)
	personID := personResult["id"].(string)

	t.Run("create donation", func(t *testing.T) {
		resp, err := ts.post("/api/giving/donations", map[string]interface{}{
			"person_id": personID,
			"fund_id":   fundID,
			"amount":    100.50,
			"method":    "cash",
			"date":      time.Now().Format("2006-01-02"),
			"notes":     "Sunday morning offering",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected id in response")
		}
		assertEqual(t, personID, result["person_id"])
		assertEqual(t, fundID, result["fund_id"])
		
		// Amount might be returned as float64
		amount := result["amount"].(float64)
		if amount < 100.49 || amount > 100.51 {
			t.Fatalf("Expected amount 100.50, got %v", amount)
		}
	})

	t.Run("list donations", func(t *testing.T) {
		// Create multiple donations
		donations := []map[string]interface{}{
			{"person_id": personID, "fund_id": fundID, "amount": 50.00, "method": "check", "date": time.Now().Format("2006-01-02")},
			{"person_id": personID, "fund_id": fundID, "amount": 75.25, "method": "online", "date": time.Now().Format("2006-01-02")},
		}

		for _, donation := range donations {
			resp, err := ts.post("/api/giving/donations", donation)
			assertNoError(t, err)
			assertStatusCode(t, resp, http.StatusCreated)
		}

		// List all donations
		resp, err := ts.get("/api/giving/donations")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if len(result) < 2 {
			t.Fatalf("Expected at least 2 donations, got %d", len(result))
		}
	})

	t.Run("list donations with date filter", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		resp, err := ts.get(fmt.Sprintf("/api/giving/donations?start_date=%s&end_date=%s", today, today))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		// Should return donations created today
		if len(result) == 0 {
			t.Fatal("Expected donations for today")
		}
	})

	t.Run("get donation by id", func(t *testing.T) {
		// Create a donation
		resp, err := ts.post("/api/giving/donations", map[string]interface{}{
			"person_id": personID,
			"fund_id":   fundID,
			"amount":    200.00,
			"method":    "cash",
			"date":      time.Now().Format("2006-01-02"),
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var createResult map[string]interface{}
		err = decodeJSON(resp, &createResult)
		assertNoError(t, err)
		donationID := createResult["id"].(string)

		// Get the donation
		getResp, err := ts.get(fmt.Sprintf("/api/giving/donations/%s", donationID))
		assertNoError(t, err)
		assertStatusCode(t, getResp, http.StatusOK)

		var getResult map[string]interface{}
		err = decodeJSON(getResp, &getResult)
		assertNoError(t, err)

		assertEqual(t, donationID, getResult["id"])
	})
}

func TestGivingStats(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create fund and person
	fundResp, err := ts.post("/api/giving/funds", map[string]interface{}{
		"name":       "General Fund",
		"is_default": true,
	})
	assertNoError(t, err)
	assertStatusCode(t, fundResp, http.StatusOK)

	var fundResult map[string]interface{}
	err = decodeJSON(fundResp, &fundResult)
	assertNoError(t, err)
	fundID := fundResult["id"].(string)

	personResp, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "John",
		"last_name":  "Donor",
	})
	assertNoError(t, err)
	assertStatusCode(t, personResp, http.StatusOK)

	var personResult map[string]interface{}
	err = decodeJSON(personResp, &personResult)
	assertNoError(t, err)
	personID := personResult["id"].(string)

	// Create donations
	donations := []float64{100.00, 200.00, 150.00}
	for _, amount := range donations {
		resp, err := ts.post("/api/giving/donations", map[string]interface{}{
			"person_id": personID,
			"fund_id":   fundID,
			"amount":    amount,
			"method":    "cash",
			"date":      time.Now().Format("2006-01-02"),
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)
	}

	t.Run("get giving stats", func(t *testing.T) {
		resp, err := ts.get("/api/giving/stats")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		// Verify total is correct (should be sum of all donations)
		if result["total"] != nil {
			total := result["total"].(float64)
			if total < 449 || total > 451 { // Allow small floating point differences
				t.Fatalf("Expected total around 450, got %v", total)
			}
		}
	})

	t.Run("get person giving history", func(t *testing.T) {
		resp, err := ts.get(fmt.Sprintf("/api/giving/person/%s", personID))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if len(result) != 3 {
			t.Fatalf("Expected 3 donations, got %d", len(result))
		}
	})
}
