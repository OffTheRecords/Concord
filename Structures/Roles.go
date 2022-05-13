package Structures

import "time"

var UNVERIFIED_ROLE = Role{ID: 1, Name: "unverified", Expiry: 0, Creation: time.Now().Unix()}
