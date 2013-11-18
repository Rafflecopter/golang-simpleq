// Lua scripts for simpleq
package scripts

import "github.com/garyburd/redigo/redis"

var SafePullPipe = redis.NewScript(
	2, // number of keys
	`local rem = redis.call("lrem", KEYS[1], -1, ARGV[1])
  if (rem ~= nil) and (rem > 0) then
    return redis.call("lpush", KEYS[2], ARGV[1])
  end
  return 0`)
