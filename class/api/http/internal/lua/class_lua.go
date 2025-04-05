package lua

// 创建课程的脚本
const CreateClassScript = `
	local classInfoKey = KEYS[1]
	local teacherClassListKey = KEYS[2]
	local classListKey = KEYS[3]

	local score = ARGV[1]
	local rawData = ARGV[2]
	local cid = ARGV[3]

	-- 设置classInfoKey的值
	redis.call("SET",classInfoKey,rawData)
	
	-- 将cid添加到classListKey的ZSET中，使用score作为分数
	redis.call("ZAdd",classListKey,score,cid)

	-- 将cid添加到teacherClassListKey的ZSET中，使用score作为分数
	redis.call("ZADD", teacherClassListKey, score, cid)

	return redis.status_reply("OK")
`

const MoveClassToTrash = `
	local teacherClassListKey = KEYS[1]
	local teacherDeleteClassListKey = KEYS[2]

	local classListKey = KEYS[3]
	local classDeleteListKey = KEYS[4]

	local cid = ARGV[1]

	local ts = ARGV[2]

	-- 从老师的课程列表中删除
	redis.call("ZRem",teacherClassListKey,cid)

	-- 添加到老师的删除课程列表中
	redis.call("ZAdd",teacherDeleteClassListKey,ts,cid)

	redis.call("ZRem",classListKey,cid)
	redis.call("ZAdd",classDeleteListKey,ts,cid)

	return redis.status_reply("OK")
`

const RecoverClass = `
	local teacherClassListKey = KEYS[1]
	local teacherDeleteClassListKey = KEYS[2]

	local classListKey = KEYS[3]
	local classDeleteListKey = KEYS[4]

	local cid = ARGV[1]

	local ts = ARGV[2]

	redis.call("ZRem",teacherDeleteClassListKey,cid)
	
	redis.call("ZAdd",teacherClassListKey,ts,cid)

	redis.call("ZRem",classDeleteListKey,cid)
	redis.call("ZAdd",classListKey,ts,cid)

	return redis.status_reply("OK")
`

const DeleteClass = `
	local teacherDeleteClassListKey = KEYS[1]
	local classDeleteListKey = KEYS[2]
	local classInfoKey = KEYS[3]

	redis.call("ZRem",teacherDeleteClassListKey,cid)
	redis.call("ZRem",classDeleteListKey,cid)
	redis.call("DEL",classInfoKey)

	 return redis.status_reply("OK")

`
