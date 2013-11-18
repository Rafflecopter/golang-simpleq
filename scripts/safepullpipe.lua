-- KEYS:[from, to], ARGV:[id]
local rem = redis.call("lrem", KEYS[1], -1, ARGV[1])
if (rem ~= nil) and (rem > 0) then
  return redis.call("lpush", KEYS[2], ARGV[1])
end
return 0