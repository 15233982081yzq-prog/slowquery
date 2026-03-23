wrk.method = "GET"
wrk.headers["Content-Type"] = "application/json"
baseUrl = "http://10.129.120.132:30083/rds/smart/v1/api/non-live/slowquery/query_detail"


local databases = {"test1","test2","test3"}
local fingerIds = {"88157ffb44b802aafd289782855d60874c1dc592915e128d32a19aed0584f0cf","db529b91f26867fc2043c6854fdfc5541de71aafed4aef9093d8ab25c2dd8efb","ba81b56f1aa5d4f5f6af225b5e1808780a272d05ac16ea8f9503a6c6eead6ef4",
                   "00ced5368727ca7537f52b2e90ffb955f55d61acb3f33f9b45e0c0e70d112cdc","b3417221843ebcbd45675703943d2ae7b4f4d6a252474b7aae05da85a9a3989d","b235ad279880ca66c3fc75fe6126c83555555d689c4f078675c0c50b05cc81a7",
                   "84fb5bc2aad9eaa5a3170a95fa24410c7dc76c5a22825a58b84424a880fe7a90","3cccdc8d9b39f9e48a8bba097bf13cdc2f457c993e014240ffe9a04ac0e22a80","2f5ce9453312303d8969a65c718f293a189dc29917d6e22e73f59868ca809a9b",
                   "c1b09234accf3f91a260578b5311d27fd7e537f00d3be08c9b34a4d2b338b2b0","e08d5e56f389b11ff6b4dcd12e82c1700860a0c15fbed87db69916f5300d269b","d48a2abe92ff6f68c587f206e3f6617205e5fe63a5f801f614a19805da2ee450",
                   "060cf7ae5752c88f47af5f84507567a0123323aefbe19303ae9cf0dd35db39e1","d0032da30275b6aa025f70e1912b66b39013995be52256a83bf4372a1666e415","fc1a3fdac953901fccb8487814f605b636089a5427dcf22e33d99d75f8d66a15",
                   "efdb9920f583caeeab530569ebe8092eaa5f2106e053e079d36384b591d096fe","103293b3bf8c1ba2bd380d4e625eef6ee2e3b9e5e5574bd3c414f11e46cf63a6","15bd6f408b013dbacfaa61dd5b8457ece525f6de3fa8cf44765f22f496ac769f",
                   "08cdffec478df644fd469fad2f70482abfe9ca87009810b38c3eba24ebe1e6aa","dbf5749b26bf3f254c726e6e1b40e9b3eef178d6d7edc5e30b8014efd2afc7cd","32ae305dae68dd07e2d7d21a1a9ad385450130be9f92d93090b4cadfea651b9b",
                   "a02e3222896ff62a2fcac4540812330519f538225c967c72ed6f24ca2d6730bf","fb8c69677b33f2fec2c700ffe8628e44cad80dfbb48571f08799a052161bc46a","7104929c29b12392a28289d006ddd181ef6652ecfb6a989672552347311d3cb2",
                   "97806864d5d8e94e6f285cc1631e743120389aaf6119160ad01fbf5587fc67ce"}
local auth = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImI3OTBlYjE5MmY2YTJkMTZjNDQwZWE3NGRlODU4ODg1NDc5NjZmZDEiLCJ0eXAiOiJKV1QifQ.eyJlbWFpbCI6ImppYW4uYmlhbkBzaG9wZWUuY29tIiwiZXhwIjoxNjkzNzkzNTAyLCJpc3MiOiJodHRwczovL3NwYWNlLnRlc3Quc2hvcGVlLmlvL3YxL2NlcnRzIiwic3ViIjoiMzE5NjExMzMifQ.GkYn889_YLY5bgUZ0ctrE-gzDJ7vrqYf106fdQWPlAA0TaGqneMV2SOMpSpInXra4bnIDHh3lwPdwfmsPnLa1d3uld5ZqRW830UFllNpiuvnTp3iKdT78WuJqYatg1akVg2Eo--rX-NE18K1vqX2cgIV5_B1W0tfcHNco-OyZHLrl_zhs6mU3tTk-pbjq5CCCpiGqJYq5Hg81LkTMKy1pYZbH24OLLn3HUMlIBDqrliwkWyB9SE8OSDpcw7xFzWl9ELy8l8ezkmXNuq8vliWQVFTFeK509ip_r9wnXr4QpaARoVCvRctr9yhEJBsJFQpEv0ChkhDlV72DThTZ_IAjxTYzKXabIUAVSvypfA9xbIEn5O9n06PGsYBKHaYHauKhjxQRPTbZKQ54txaa4bNB8LZncMwEFPzD7NFdYU0Xoo5PxkiU39fnOqc6nj7k9XHkjqjkw_L-_qf9KKoJ181jklwipgihDyeImilvY4V776tB3HLyOQKwgQnQr77_voi6B7WZDRWtSb-9LnCucnXoS0YwiXaI8h2C4jhNJfC2CN_JsYVox0SIQNzcgveG9l4Y2-PsPOsdsY0Jrl0eYaYeQKkGvXctxM7ffraU7NEjh-Abx4EvuP1C5YSA5x5XY0yftfCiUWCFcRxKeMhEB8L6aonPLGr_lZrxylROz1rYqg"
local env = "pfb"
local params = {}

function request()
    wrk.headers["Authorization"] = "Bearer " .. auth
    params = build_query_map()
    queryStr = build_query_string(params)
    local full_url = baseUrl .. "?" .. queryStr-- 去除最后一个"&"


    -- print("Generated query_detail URL: " .. full_url)  -- 打印生成的URL
    return wrk.format(nil, full_url)
end



function build_query_map()
    local start_timestamp = build_start_timestamp()
    local end_timestamp = random_time_range(start_timestamp)

    params["db_name"] = random_database(databases)
    params["db_env"] = env
    params["finger_id"] = random_fingerId(fingerIds)
    params["start_time"] = start_timestamp
    params["end_time"] = end_timestamp

    return params
end

function random_fingerId(fingerIds)
    local random_index = math.random(#fingerIds)
    return fingerIds[random_index]
end

-- 随机选择数据库函数
function random_database(databases)
    local random_index = math.random(#databases)
    return databases[random_index]
end

-- 请求参数拼装函数
function build_query_string(params)
    local query_string = ""
    for key, value in pairs(params) do
        query_string = query_string .. key .. "=" .. value .. "&"
    end
    return query_string:sub(1, -2) -- 去除最后一个"&"
end


-- 获取当前日期和时间
local current_time = os.time()

-- 获取当前日期的年、月、日部分
local current_date_table = os.date("*t", current_time)
local current_year = current_date_table.year
local current_month = current_date_table.month
local current_day = current_date_table.day


-- 构建指定时间的日期时间表
local target_time_table = {
    year = current_year,
    month = current_month,
    day = current_day,
    hour = 0,
    min = 0,
    sec = 1
}

function build_start_timestamp()
    return os.time(target_time_table)
end

-- 生成12-23小时之间的随机时间戳函数
function random_time_range(base_timestamp)
    local offset = math.random(1 * 3600, 23 * 3600) -- 1-23小时转换为秒
    return base_timestamp + offset
end