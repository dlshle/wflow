package utils

import (
	"math/rand"
	"time"

	"github.com/gofrs/uuid"
)

func RandomUUID() string {
	v4UUID, err := uuid.NewV4()
	if err == nil {
		return v4UUID.String()
	}
	return generateUUIDRaw()
}

/*
function generateUUID() { // Public Domain/MIT

		var d = new Date().getTime();//Timestamp
		var d2 = ((typeof performance !== 'undefined') && performance.now && (performance.now()*1000)) || 0;//Time in microseconds since page-load or 0 if unsupported
		return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
			var r = Math.random() * 16;//random number between 0 and 16
			if(d > 0){//Use timestamp until depleted
				r = (d + r)%16 | 0;
				d = Math.floor(d/16);
			} else {//Use microseconds since page-load if supported
				r = (d2 + r)%16 | 0;
				d2 = Math.floor(d2/16);
			}
			return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
		});
	}
*/
func generateUUIDRaw() string {
	t := int(time.Now().UnixMilli())
	t2 := int(time.Now().UnixMilli())
	uuidBytes := []byte("xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx")
	for i, c := range uuidBytes {
		r := rand.Intn(16)
		if t > 0 {
			r = (t + r) % 16
			t = t / 16
		} else {
			r = (t2 + r) % 16
			t2 = t2 / 16
		}
		if c == 'x' {
			uuidBytes[i] = byte(r)
		} else {
			uuidBytes[i] = byte(r&0x3 | 0x8)
		}
	}
	return string(uuidBytes)
}
