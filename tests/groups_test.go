package tests

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGroupsCreate(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	t.Run("create group successfully", func(t *testing.T) {
		resp, err := ts.post("/api/groups", map[string]interface{}{
			"name":        "Small Group Alpha",
			"description": "Tuesday evening small group",
			"type":        "small_group",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected id in response")
		}
		assertEqual(t, "Small Group Alpha", result["name"])
		assertEqual(t, "Tuesday evening small group", result["description"])
		// Type field is optional
		if result["type"] != nil {
			assertEqual(t, "small_group", result["type"])
		}
	})

	t.Run("create group with minimal fields", func(t *testing.T) {
		resp, err := ts.post("/api/groups", map[string]interface{}{
			"name": "Youth Group",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, "Youth Group", result["name"])
	})
}

func TestGroupsList(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create test groups
	groups := []map[string]interface{}{
		{"name": "Men's Group", "type": "ministry"},
		{"name": "Women's Group", "type": "ministry"},
		{"name": "Youth Group", "type": "ministry"},
	}

	for _, group := range groups {
		resp, err := ts.post("/api/groups", group)
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)
	}

	t.Run("list all groups", func(t *testing.T) {
		resp, err := ts.get("/api/groups")
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if len(result) != 3 {
			t.Fatalf("Expected 3 groups, got %d", len(result))
		}
	})
}

func TestGroupsGetUpdate(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a group
	resp, err := ts.post("/api/groups", map[string]interface{}{
		"name":        "Small Group",
		"description": "Original description",
	})
	assertNoError(t, err)
	assertStatusCode(t, resp, http.StatusCreated)

	var createResult map[string]interface{}
	err = decodeJSON(resp, &createResult)
	assertNoError(t, err)
	groupID := createResult["id"].(string)

	t.Run("get group by id", func(t *testing.T) {
		resp, err := ts.get(fmt.Sprintf("/api/groups/%s", groupID))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, groupID, result["id"])
		assertEqual(t, "Small Group", result["name"])
	})

	t.Run("update group", func(t *testing.T) {
		resp, err := ts.put(fmt.Sprintf("/api/groups/%s", groupID), map[string]interface{}{
			"name":        "Updated Group Name",
			"description": "Updated description",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		assertEqual(t, "Updated Group Name", result["name"])
		assertEqual(t, "Updated description", result["description"])
	})
}

func TestGroupsDelete(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a group
	resp, err := ts.post("/api/groups", map[string]interface{}{
		"name": "Temporary Group",
	})
	assertNoError(t, err)
	assertStatusCode(t, resp, http.StatusCreated)

	var createResult map[string]interface{}
	err = decodeJSON(resp, &createResult)
	assertNoError(t, err)
	groupID := createResult["id"].(string)

	t.Run("delete group", func(t *testing.T) {
		resp, err := ts.delete(fmt.Sprintf("/api/groups/%s", groupID))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		// Verify group is deleted
		getResp, err := ts.get(fmt.Sprintf("/api/groups/%s", groupID))
		assertNoError(t, err)
		assertStatusCode(t, getResp, http.StatusNotFound)
	})
}

func TestGroupMembers(t *testing.T) {
	cleanupBeforeTest(t)
	ts := newTestServer()
	defer ts.Close()

	token, err := ts.register("Test Church", "admin@testchurch.com", "password123")
	assertNoError(t, err)
	ts.setAuthToken(token)

	// Create a group
	groupResp, err := ts.post("/api/groups", map[string]interface{}{
		"name": "Small Group",
	})
	assertNoError(t, err)
	assertStatusCode(t, groupResp, http.StatusOK)

	var groupResult map[string]interface{}
	err = decodeJSON(groupResp, &groupResult)
	assertNoError(t, err)
	groupID := groupResult["id"].(string)

	// Create people
	personResp1, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "John",
		"last_name":  "Member",
	})
	assertNoError(t, err)
	assertStatusCode(t, personResp1, http.StatusOK)

	var personResult1 map[string]interface{}
	err = decodeJSON(personResp1, &personResult1)
	assertNoError(t, err)
	personID1 := personResult1["id"].(string)

	personResp2, err := ts.post("/api/people", map[string]interface{}{
		"first_name": "Jane",
		"last_name":  "Member",
	})
	assertNoError(t, err)
	assertStatusCode(t, personResp2, http.StatusOK)

	var personResult2 map[string]interface{}
	err = decodeJSON(personResp2, &personResult2)
	assertNoError(t, err)
	personID2 := personResult2["id"].(string)

	t.Run("add member to group", func(t *testing.T) {
		resp, err := ts.post(fmt.Sprintf("/api/groups/%s/members", groupID), map[string]interface{}{
			"person_id": personID1,
			"role":      "leader",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if result["id"] == nil {
			t.Fatal("Expected membership id in response")
		}
		assertEqual(t, personID1, result["person_id"])
		assertEqual(t, "leader", result["role"])
	})

	t.Run("add another member", func(t *testing.T) {
		resp, err := ts.post(fmt.Sprintf("/api/groups/%s/members", groupID), map[string]interface{}{
			"person_id": personID2,
			"role":      "member",
		})
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)
	})

	t.Run("list group members", func(t *testing.T) {
		resp, err := ts.get(fmt.Sprintf("/api/groups/%s/members", groupID))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		if len(result) != 2 {
			t.Fatalf("Expected 2 members, got %d", len(result))
		}
	})

	t.Run("update member role", func(t *testing.T) {
		// Get the member ID first
		resp, err := ts.get(fmt.Sprintf("/api/groups/%s/members", groupID))
		assertNoError(t, err)
		
		var members []interface{}
		err = decodeJSON(resp, &members)
		assertNoError(t, err)

		member := members[0].(map[string]interface{})
		memberID := member["id"].(string)

		// Update the role
		updateResp, err := ts.put(fmt.Sprintf("/api/groups/%s/members/%s", groupID, memberID), map[string]interface{}{
			"role": "co-leader",
		})
		assertNoError(t, err)
		assertStatusCode(t, updateResp, http.StatusOK)

		var result map[string]interface{}
		err = decodeJSON(updateResp, &result)
		assertNoError(t, err)

		assertEqual(t, "co-leader", result["role"])
	})

	t.Run("remove member from group", func(t *testing.T) {
		// Get the member ID first
		resp, err := ts.get(fmt.Sprintf("/api/groups/%s/members", groupID))
		assertNoError(t, err)
		
		var members []interface{}
		err = decodeJSON(resp, &members)
		assertNoError(t, err)

		member := members[0].(map[string]interface{})
		memberID := member["id"].(string)

		// Remove the member
		deleteResp, err := ts.delete(fmt.Sprintf("/api/groups/%s/members/%s", groupID, memberID))
		assertNoError(t, err)
		assertStatusCode(t, deleteResp, http.StatusOK)

		// Verify member was removed
		listResp, err := ts.get(fmt.Sprintf("/api/groups/%s/members", groupID))
		assertNoError(t, err)
		
		var updatedMembers []interface{}
		err = decodeJSON(listResp, &updatedMembers)
		assertNoError(t, err)

		if len(updatedMembers) != 1 {
			t.Fatalf("Expected 1 member after removal, got %d", len(updatedMembers))
		}
	})

	t.Run("get person groups", func(t *testing.T) {
		resp, err := ts.get(fmt.Sprintf("/api/groups/person/%s", personID2))
		assertNoError(t, err)
		assertStatusCode(t, resp, http.StatusCreated)

		var result []interface{}
		err = decodeJSON(resp, &result)
		assertNoError(t, err)

		// personID2 should still be in the group
		if len(result) != 1 {
			t.Fatalf("Expected 1 group for person, got %d", len(result))
		}
	})
}
