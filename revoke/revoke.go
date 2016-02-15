package revoke

import (
	"encoding/json"
	"log"
	"time"
)

type Revocation struct {
	Type      string // token, claim, global
	Data      map[string]interface{}
	Timestamp int
}

type jsonRevoke struct {
	Meta struct {
		ForceRefresh string `json:"force_refresh"`
	} `json:"meta"`
	Revs []jsonRevocation `json:"revocations"`
}

type jsonRevocation struct {
	Type      string `json:"type"` // TOKEN, CLAIM, GLOBAL
	RevokedAt int    `json:"revoked_at"`
	Data      struct {
		Name          string `json:"name,omitempty"`           // CLAIM
		ValueHash     string `json:"value_hash,omitempty"`     // CLAIM
		IssuedBefore  int    `json:"issued_before,omitempty"`  // CLAIM, GLOBAL
		TokenHash     string `json:"token_hash,omitempty"`     // TOKEN
		HashAlgorithm string `json:"hash_algorithm,omitempty"` // CLAIM, TOKEN
	} `json:"data"`
}

func (r *jsonRevoke) UnmarshallJSON(data []byte) (err error) {
	if err = json.Unmarshal(data, &r); err != nil {
		log.Println("Error unmarshalling revocation json. " + err.Error())
		return err
	}

	// Note: if we already foreced a refresh, we don't want to do it again
	// otherwise we'll get stuck in an infinite loop
	/*
		if buf.ForceRefresh != "" && !forcedRefresh {
			i, err := strconv.Atoi(buf.ForceRefresh)
			if err != nil {
				log.Println("Error converting ForceRefresh to int." + err.Error())
			} else {
				// TODO: not sure how to get the current Cache from here. . .
				refreshCacheFromTime(i)
			}
		}
	*/

	return
}

func (r *Revocation) getRevocationFromJson(j *jsonRevocation) {

	// account for some network delay, say three seconds
	t := int(time.Now().UnixNano()/1e6) - 3*1e6

	r.Data = make(map[string]interface{})
	switch j.Type {
	case "TOKEN":
		valid := isHashTimestampValid(j.Data.TokenHash, j.RevokedAt)
		if !valid {
			log.Println("Invalid revocation data (TOKEN). TokenHash: %s, RevokedAt: %d", j.Data.TokenHash, j.RevokedAt)
			return
		}
		r.Data["token_hash"] = j.Data.TokenHash
		r.Data["revoked_at"] = j.RevokedAt
	case "CLAIM":
		valid := isHashTimestampValid(j.Data.ValueHash, j.Data.IssuedBefore)
		if !valid {
			log.Println("Invalid revocation data (CLAIM). ValueHash: %s, IssuedBefore: %d", j.Data.ValueHash, j.Data.IssuedBefore)
			return
		}
		if j.Data.Name == "" {
			log.Println("Invalid revocation data (missing claim name).")
			return
		}
		r.Data["value_hash"] = j.Data.ValueHash
		r.Data["issued_before"] = j.Data.IssuedBefore
		r.Data["name"] = j.Data.Name
	case "GLOBAL":
		valid := isHashTimestampValid("thisStringDoesntMatter", j.Data.IssuedBefore)
		if !valid {
			log.Println("Invalid revocation data (GLOBAL). IssuedBefore: %d", j.Data.IssuedBefore)
			return
		}
		r.Data["issued_before"] = j.Data.IssuedBefore
	default:
		log.Println("Unsupported revocation type: %s", j.Type)
		return
	}

	r.Type = j.Type
	r.Timestamp = t
	return
}

func isHashTimestampValid(hash string, timestamp int) bool {
	return hash != "" && timestamp != 0
}

// vim: ts=4 sw=4 noexpandtab nolist syn=go