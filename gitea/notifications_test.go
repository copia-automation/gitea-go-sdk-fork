// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gitea

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotifications(t *testing.T) {
	log.Println("== TestNotifications ==")

	// init user2
	c := newTestClient()

	user1, _, err := c.GetMyUserInfo()
	assert.NoError(t, err)
	user2 := createTestUser(t, "notify2", c)

	//create 2 repos
	repoA, err := createTestRepo(t, "TestNotifications_A", c)
	assert.NoError(t, err)

	c.sudo = user2.UserName
	repoB, err := createTestRepo(t, "TestNotifications_B", c)
	assert.NoError(t, err)
	_, err = c.WatchRepo(user1.UserName, repoA.Name)
	c.sudo = ""
	assert.NoError(t, err)

	c.sudo = user2.UserName
	_, err = c.ReadNotifications(MarkNotificationOptions{})
	assert.NoError(t, err)
	count, _, err := c.CheckNotifications()
	assert.EqualValues(t, 0, count)
	assert.NoError(t, err)
	c.sudo = ""
	_, _, err = c.CreateIssue(repoA.Owner.UserName, repoA.Name, CreateIssueOption{Title: "A Issue", Closed: false})
	assert.NoError(t, err)
	issue, _, err := c.CreateIssue(repoB.Owner.UserName, repoB.Name, CreateIssueOption{Title: "B Issue", Closed: false})
	assert.NoError(t, err)
	time.Sleep(time.Second * 1)

	// CheckNotifications of user2
	c.sudo = user2.UserName
	count, _, err = c.CheckNotifications()
	assert.NoError(t, err)
	assert.EqualValues(t, 2, count)

	// ListNotifications
	nList, _, err := c.ListNotifications(ListNotificationOptions{})
	assert.NoError(t, err)
	assert.Len(t, nList, 2)
	for _, n := range nList {
		assert.EqualValues(t, true, n.Unread)
		assert.EqualValues(t, "Issue", n.Subject.Type)
		if n.Subject.Title == "A Issue" {
			assert.EqualValues(t, repoA.Name, n.Repository.Name)
		} else if n.Subject.Title == "B Issue" {
			assert.EqualValues(t, repoB.Name, n.Repository.Name)
		} else {
			assert.Error(t, fmt.Errorf("ListNotifications returned a Issue witch should not"))
		}
	}

	// ListRepoNotifications
	nList, _, err = c.ListRepoNotifications(repoA.Owner.UserName, repoA.Name, ListNotificationOptions{})
	assert.NoError(t, err)
	assert.Len(t, nList, 1)
	assert.EqualValues(t, "A Issue", nList[0].Subject.Title)
	// ReadRepoNotifications
	_, err = c.ReadRepoNotifications(repoA.Owner.UserName, repoA.Name, MarkNotificationOptions{})
	assert.NoError(t, err)

	// GetThread
	n, _, err := c.GetNotification(nList[0].ID)
	assert.NoError(t, err)
	assert.EqualValues(t, false, n.Unread)
	assert.EqualValues(t, "A Issue", n.Subject.Title)

	// ReadNotifications
	_, err = c.ReadNotifications(MarkNotificationOptions{})
	assert.NoError(t, err)
	nList, _, err = c.ListNotifications(ListNotificationOptions{})
	assert.NoError(t, err)
	assert.Len(t, nList, 0)

	// ReadThread
	iState := StateClosed
	c.sudo = ""
	_, _, err = c.EditIssue(repoB.Owner.UserName, repoB.Name, issue.Index, EditIssueOption{State: &iState})
	assert.NoError(t, err)
	time.Sleep(time.Second * 1)

	c.sudo = user2.UserName
	nList, _, err = c.ListNotifications(ListNotificationOptions{})
	assert.NoError(t, err)
	count, _, err = c.CheckNotifications()
	assert.NoError(t, err)
	assert.EqualValues(t, 1, count)
	assert.Len(t, nList, 1)
	if len(nList) > 0 {
		_, err = c.ReadNotification(nList[0].ID)
		assert.NoError(t, err)
	}

	c.sudo = ""
	_, err = c.ReadNotifications(MarkNotificationOptions{})
	assert.NoError(t, err)
	_, _ = c.DeleteRepo("test01", "Reviews")
	nList, _, err = c.ListNotifications(ListNotificationOptions{Status: []NotifyStatus{NotifyStatusRead}})
	assert.NoError(t, err)
	assert.Len(t, nList, 4)

	_, err = c.ReadNotification(nList[2].ID, NotifyStatusPinned)
	assert.NoError(t, err)
	_, err = c.ReadNotification(nList[3].ID, NotifyStatusUnread)
	assert.NoError(t, err)
	nList, _, err = c.ListNotifications(ListNotificationOptions{Status: []NotifyStatus{NotifyStatusPinned, NotifyStatusUnread}})
	assert.NoError(t, err)
	assert.Len(t, nList, 2)
}
