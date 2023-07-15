package nostrfy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/boltcard/boltcard/db"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	log "github.com/sirupsen/logrus"
)

func SendNostrfication(card_id int, card_name string, payOrRec int) {
	requestURL := fmt.Sprintf("https://%s/.well-known/nostr.json?name=%s", db.Get_setting("HOST_DOMAIN"), card_name)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warn(err.Error())
	}
	nipNames := make(map[string]interface{})
	if strings.Contains(string(resBody), "names") {
		var res map[string]any
		if err := json.Unmarshal([]byte(resBody), &res); err != nil {
			log.Warn(err.Error())
			return
		}
		nnames := res["names"].(map[string]any)
		for key, value := range nnames {
			nipNames[key] = value.(string)
		}
	} else {
		log.Warn("No NIP-05 addresses...Unable to send payment/receipt info")
		return
	}

	bolt_bot_privkey := db.Get_setting("NOSTR_BOT_PRIVKEY_HEX")
	nostr_rel_list := db.Get_setting("NOSTR_RELAYS_LIST")
	if bolt_bot_privkey != "" && nostr_rel_list != "" && len(nipNames) != 0 {
		last_tnx, err := db.Get_latest_card_tx(card_id, payOrRec)
		if err != nil {
			log.Warn(err.Error())
			return
		}
		nmsg := fmt.Sprintf("you made payment of %d sats via BoltCard", last_tnx.Tx_amount_msats/1000)
		if payOrRec == db.NostrRec {
			nmsg = fmt.Sprintf("you received %d sats via BoltCard service", last_tnx.Tx_amount_msats/1000)
		}
		tags := make(nostr.Tags, 0)
		for nkey, npub := range nipNames {
			msg := fmt.Sprintf("Hey %s, %s - %s", nkey, nmsg, last_tnx.Tx_time)
			pub, _ := nostr.GetPublicKey(bolt_bot_privkey)
			css, _ := nip04.ComputeSharedSecret(npub.(string), bolt_bot_privkey)
			emsg, _ := nip04.Encrypt(msg, css)
			tag1 := nostr.Tag{"p", npub.(string)}
			tt := tags.AppendUnique(tag1)
			ev := nostr.Event{
				PubKey:    pub,
				CreatedAt: nostr.Now(),
				Kind:      4,
				Tags:      tt,
				Content:   emsg,
			}
			// calling Sign sets the event ID field and the event Sig field
			ev.Sign(bolt_bot_privkey)
			for _, url := range []string{nostr_rel_list} {
				ctx := context.WithValue(context.Background(), "url", url)
				relay, err := nostr.RelayConnect(ctx, url)
				if err != nil {
					log.Warn(err.Error())
					continue
				}
				_, err = relay.Publish(ctx, ev)
				if err != nil {
					log.Warn(err.Error())
					continue
				}
				log.WithFields(log.Fields{"published to ": url}).Info("payment/reciept message sent")
			}
		}
	}
}
