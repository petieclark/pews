package tests

import (
	"fmt"
	"net/http"
	"testing"
)

// TestMultiTenantIsolation is a CRITICAL test that verifies data is properly isolated between tenants
func TestMultiTenantIsolation(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	// Create Tenant A
	tokenA, err := ts.register("Church Alpha", "admin@alpha.com", "password123")
	assertNoError(t, err)
	if tokenA == "" {
		t.Fatal("Failed to create tenant A")
	}

	// Create Tenant B
	tokenB, err := ts.register("Church Beta", "admin@beta.com", "password123")
	assertNoError(t, err)
	if tokenB == "" {
		t.Fatal("Failed to create tenant B")
	}

	t.Run("people isolation", func(t *testing.T) {
		// Tenant A: Create a person
		ts.setAuthToken(tokenA)
		respA, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "Alice",
			"last_name":  "Alpha",
			"email":      "alice@alpha.com",
		})
		assertNoError(t, err)
		assertStatusCode(t, respA, http.StatusOK)

		var personA map[string]interface{}
		err = decodeJSON(respA, &personA)
		assertNoError(t, err)
		personAID := personA["id"].(string)

		// Tenant B: Create a person
		ts.setAuthToken(tokenB)
		respB, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "Bob",
			"last_name":  "Beta",
			"email":      "bob@beta.com",
		})
		assertNoError(t, err)
		assertStatusCode(t, respB, http.StatusOK)

		var personB map[string]interface{}
		err = decodeJSON(respB, &personB)
		assertNoError(t, err)

		// Verify Tenant A can see only their person
		ts.setAuthToken(tokenA)
		listRespA, err := ts.get("/api/people")
		assertNoError(t, err)
		assertStatusCode(t, listRespA, http.StatusOK)

		var listResultA map[string]interface{}
		err = decodeJSON(listRespA, &listResultA)
		assertNoError(t, err)

		peopleA := listResultA["people"].([]interface{})
		if len(peopleA) != 1 {
			t.Fatalf("Tenant A should see 1 person, got %d", len(peopleA))
		}

		personInListA := peopleA[0].(map[string]interface{})
		if personInListA["first_name"] != "Alice" {
			t.Fatalf("Tenant A should see Alice, got %v", personInListA["first_name"])
		}

		// Verify Tenant B can see only their person
		ts.setAuthToken(tokenB)
		listRespB, err := ts.get("/api/people")
		assertNoError(t, err)
		assertStatusCode(t, listRespB, http.StatusOK)

		var listResultB map[string]interface{}
		err = decodeJSON(listRespB, &listResultB)
		assertNoError(t, err)

		peopleB := listResultB["people"].([]interface{})
		if len(peopleB) != 1 {
			t.Fatalf("Tenant B should see 1 person, got %d", len(peopleB))
		}

		personInListB := peopleB[0].(map[string]interface{})
		if personInListB["first_name"] != "Bob" {
			t.Fatalf("Tenant B should see Bob, got %v", personInListB["first_name"])
		}

		// CRITICAL: Verify Tenant B cannot access Tenant A's person by ID
		ts.setAuthToken(tokenB)
		getResp, err := ts.get(fmt.Sprintf("/api/people/%s", personAID))
		assertNoError(t, err)
		if getResp.StatusCode != http.StatusNotFound && getResp.StatusCode != http.StatusForbidden {
			t.Fatalf("Tenant B should not be able to access Tenant A's person, got status %d", getResp.StatusCode)
		}
	})

	t.Run("groups isolation", func(t *testing.T) {
		// Tenant A: Create a group
		ts.setAuthToken(tokenA)
		respA, err := ts.post("/api/groups", map[string]interface{}{
			"name": "Alpha Group",
		})
		assertNoError(t, err)
		assertStatusCode(t, respA, http.StatusOK)

		var groupA map[string]interface{}
		err = decodeJSON(respA, &groupA)
		assertNoError(t, err)
		groupAID := groupA["id"].(string)

		// Tenant B: Create a group
		ts.setAuthToken(tokenB)
		respB, err := ts.post("/api/groups", map[string]interface{}{
			"name": "Beta Group",
		})
		assertNoError(t, err)
		assertStatusCode(t, respB, http.StatusOK)

		// Verify Tenant A sees only their group
		ts.setAuthToken(tokenA)
		listRespA, err := ts.get("/api/groups")
		assertNoError(t, err)
		assertStatusCode(t, listRespA, http.StatusOK)

		var groupsA []interface{}
		err = decodeJSON(listRespA, &groupsA)
		assertNoError(t, err)

		if len(groupsA) != 1 {
			t.Fatalf("Tenant A should see 1 group, got %d", len(groupsA))
		}

		// Verify Tenant B sees only their group
		ts.setAuthToken(tokenB)
		listRespB, err := ts.get("/api/groups")
		assertNoError(t, err)
		assertStatusCode(t, listRespB, http.StatusOK)

		var groupsB []interface{}
		err = decodeJSON(listRespB, &groupsB)
		assertNoError(t, err)

		if len(groupsB) != 1 {
			t.Fatalf("Tenant B should see 1 group, got %d", len(groupsB))
		}

		// CRITICAL: Verify Tenant B cannot access Tenant A's group by ID
		ts.setAuthToken(tokenB)
		getResp, err := ts.get(fmt.Sprintf("/api/groups/%s", groupAID))
		assertNoError(t, err)
		if getResp.StatusCode != http.StatusNotFound && getResp.StatusCode != http.StatusForbidden {
			t.Fatalf("Tenant B should not be able to access Tenant A's group, got status %d", getResp.StatusCode)
		}
	})

	t.Run("giving isolation", func(t *testing.T) {
		// Tenant A: Create fund and donation
		ts.setAuthToken(tokenA)

		// Create person for tenant A
		personRespA, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "Donor",
			"last_name":  "Alpha",
		})
		assertNoError(t, err)
		assertStatusCode(t, personRespA, http.StatusOK)

		var personA map[string]interface{}
		err = decodeJSON(personRespA, &personA)
		assertNoError(t, err)
		personAID := personA["id"].(string)

		fundRespA, err := ts.post("/api/giving/funds", map[string]interface{}{
			"name":       "Alpha Fund",
			"is_default": true,
		})
		assertNoError(t, err)
		assertStatusCode(t, fundRespA, http.StatusOK)

		var fundA map[string]interface{}
		err = decodeJSON(fundRespA, &fundA)
		assertNoError(t, err)
		fundAID := fundA["id"].(string)

		donationRespA, err := ts.post("/api/giving/donations", map[string]interface{}{
			"person_id": personAID,
			"fund_id":   fundAID,
			"amount":    100.00,
			"method":    "cash",
			"date":      "2024-01-01",
		})
		assertNoError(t, err)
		assertStatusCode(t, donationRespA, http.StatusOK)

		// Tenant B: Create fund and donation
		ts.setAuthToken(tokenB)

		// Create person for tenant B
		personRespB, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "Donor",
			"last_name":  "Beta",
		})
		assertNoError(t, err)
		assertStatusCode(t, personRespB, http.StatusOK)

		var personB map[string]interface{}
		err = decodeJSON(personRespB, &personB)
		assertNoError(t, err)
		personBID := personB["id"].(string)

		fundRespB, err := ts.post("/api/giving/funds", map[string]interface{}{
			"name":       "Beta Fund",
			"is_default": true,
		})
		assertNoError(t, err)
		assertStatusCode(t, fundRespB, http.StatusOK)

		var fundB map[string]interface{}
		err = decodeJSON(fundRespB, &fundB)
		assertNoError(t, err)
		fundBID := fundB["id"].(string)

		donationRespB, err := ts.post("/api/giving/donations", map[string]interface{}{
			"person_id": personBID,
			"fund_id":   fundBID,
			"amount":    200.00,
			"method":    "check",
			"date":      "2024-01-01",
		})
		assertNoError(t, err)
		assertStatusCode(t, donationRespB, http.StatusOK)

		// Verify Tenant A sees only their funds
		ts.setAuthToken(tokenA)
		fundsRespA, err := ts.get("/api/giving/funds")
		assertNoError(t, err)
		assertStatusCode(t, fundsRespA, http.StatusOK)

		var fundsA []interface{}
		err = decodeJSON(fundsRespA, &fundsA)
		assertNoError(t, err)

		if len(fundsA) != 1 {
			t.Fatalf("Tenant A should see 1 fund, got %d", len(fundsA))
		}

		fundInListA := fundsA[0].(map[string]interface{})
		if fundInListA["name"] != "Alpha Fund" {
			t.Fatalf("Tenant A should see Alpha Fund, got %v", fundInListA["name"])
		}

		// Verify Tenant A sees only their donations
		donationsRespA, err := ts.get("/api/giving/donations")
		assertNoError(t, err)
		assertStatusCode(t, donationsRespA, http.StatusOK)

		var donationsA []interface{}
		err = decodeJSON(donationsRespA, &donationsA)
		assertNoError(t, err)

		if len(donationsA) != 1 {
			t.Fatalf("Tenant A should see 1 donation, got %d", len(donationsA))
		}

		donationInListA := donationsA[0].(map[string]interface{})
		amount := donationInListA["amount"].(float64)
		if amount < 99.99 || amount > 100.01 {
			t.Fatalf("Tenant A should see $100 donation, got %v", amount)
		}

		// Verify Tenant B sees only their funds
		ts.setAuthToken(tokenB)
		fundsRespB, err := ts.get("/api/giving/funds")
		assertNoError(t, err)
		assertStatusCode(t, fundsRespB, http.StatusOK)

		var fundsB []interface{}
		err = decodeJSON(fundsRespB, &fundsB)
		assertNoError(t, err)

		if len(fundsB) != 1 {
			t.Fatalf("Tenant B should see 1 fund, got %d", len(fundsB))
		}

		// CRITICAL: Verify Tenant B cannot create donation with Tenant A's fund
		ts.setAuthToken(tokenB)
		invalidDonationResp, err := ts.post("/api/giving/donations", map[string]interface{}{
			"person_id": personBID,
			"fund_id":   fundAID, // Attempting to use Tenant A's fund
			"amount":    50.00,
			"method":    "cash",
			"date":      "2024-01-01",
		})
		assertNoError(t, err)
		// Should fail with 400 or 404
		if invalidDonationResp.StatusCode == http.StatusOK {
			t.Fatal("Tenant B should not be able to create donation with Tenant A's fund")
		}
	})

	t.Run("check-ins isolation", func(t *testing.T) {
		// Tenant A: Create event and check-in
		ts.setAuthToken(tokenA)

		eventRespA, err := ts.post("/api/checkins/events", map[string]interface{}{
			"name":       "Alpha Service",
			"start_time": "2024-01-01T10:00:00Z",
			"end_time":   "2024-01-01T12:00:00Z",
		})
		assertNoError(t, err)
		assertStatusCode(t, eventRespA, http.StatusOK)

		var eventA map[string]interface{}
		err = decodeJSON(eventRespA, &eventA)
		assertNoError(t, err)
		eventAID := eventA["id"].(string)

		// Tenant B: Create event
		ts.setAuthToken(tokenB)

		eventRespB, err := ts.post("/api/checkins/events", map[string]interface{}{
			"name":       "Beta Service",
			"start_time": "2024-01-01T10:00:00Z",
			"end_time":   "2024-01-01T12:00:00Z",
		})
		assertNoError(t, err)
		assertStatusCode(t, eventRespB, http.StatusOK)

		// Verify Tenant A sees only their event
		ts.setAuthToken(tokenA)
		eventsRespA, err := ts.get("/api/checkins/events")
		assertNoError(t, err)
		assertStatusCode(t, eventsRespA, http.StatusOK)

		var eventsA []interface{}
		err = decodeJSON(eventsRespA, &eventsA)
		assertNoError(t, err)

		if len(eventsA) != 1 {
			t.Fatalf("Tenant A should see 1 event, got %d", len(eventsA))
		}

		// Verify Tenant B sees only their event
		ts.setAuthToken(tokenB)
		eventsRespB, err := ts.get("/api/checkins/events")
		assertNoError(t, err)
		assertStatusCode(t, eventsRespB, http.StatusOK)

		var eventsB []interface{}
		err = decodeJSON(eventsRespB, &eventsB)
		assertNoError(t, err)

		if len(eventsB) != 1 {
			t.Fatalf("Tenant B should see 1 event, got %d", len(eventsB))
		}

		// CRITICAL: Verify Tenant B cannot access Tenant A's event
		ts.setAuthToken(tokenB)
		getResp, err := ts.get(fmt.Sprintf("/api/checkins/events/%s", eventAID))
		assertNoError(t, err)
		if getResp.StatusCode != http.StatusNotFound && getResp.StatusCode != http.StatusForbidden {
			t.Fatalf("Tenant B should not be able to access Tenant A's event, got status %d", getResp.StatusCode)
		}
	})

	t.Run("cross-tenant update protection", func(t *testing.T) {
		// Tenant A: Create a person
		ts.setAuthToken(tokenA)
		respA, err := ts.post("/api/people", map[string]interface{}{
			"first_name": "Protected",
			"last_name":  "Person",
		})
		assertNoError(t, err)
		assertStatusCode(t, respA, http.StatusOK)

		var personA map[string]interface{}
		err = decodeJSON(respA, &personA)
		assertNoError(t, err)
		personAID := personA["id"].(string)

		// CRITICAL: Tenant B attempts to update Tenant A's person
		ts.setAuthToken(tokenB)
		updateResp, err := ts.put(fmt.Sprintf("/api/people/%s", personAID), map[string]interface{}{
			"first_name": "Hacked",
			"last_name":  "Name",
		})
		assertNoError(t, err)

		// Should fail with 404 or 403
		if updateResp.StatusCode == http.StatusOK {
			t.Fatal("Tenant B should not be able to update Tenant A's person")
		}

		// CRITICAL: Tenant B attempts to delete Tenant A's person
		deleteResp, err := ts.delete(fmt.Sprintf("/api/people/%s", personAID))
		assertNoError(t, err)

		// Should fail with 404 or 403
		if deleteResp.StatusCode == http.StatusOK {
			t.Fatal("Tenant B should not be able to delete Tenant A's person")
		}

		// Verify person is still intact for Tenant A
		ts.setAuthToken(tokenA)
		getResp, err := ts.get(fmt.Sprintf("/api/people/%s", personAID))
		assertNoError(t, err)
		assertStatusCode(t, getResp, http.StatusOK)

		var verifyPerson map[string]interface{}
		err = decodeJSON(getResp, &verifyPerson)
		assertNoError(t, err)

		assertEqual(t, "Protected", verifyPerson["first_name"])
	})
}
