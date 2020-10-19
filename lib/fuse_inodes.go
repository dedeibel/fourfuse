package fourfuse

import (
	"hash/fnv"
)

const (
	HASH_BOARD_PREFIX            = "bo"
	HASH_THREAD_PREFIX           = "th"
	HASH_THREAD_INFO_PREFIX      = "ti"
	HASH_THREAD_DISCUSSION       = "td"
	HASH_POST_PREFIX             = "po"
	HASH_POST_INFO_PREFIX        = "pi"
	HASH_POST_IMAGE              = "pm"
	HASH_POST_THUMBNAIL          = "pt"
	HASH_POST_PREFIXED_THUMBNAIL = "pp"
	HASH_POST_COMMENT            = "pc"
	HASH_POST_SUBJECT            = "ps"
	HASH_IMAGE_DIR_PREFIX        = "id"
	HASH_POSTS_DIR_PREFIX        = "pd"
)

func hashs(str string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}
