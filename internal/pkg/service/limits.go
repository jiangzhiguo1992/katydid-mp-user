package service

type (
	Limits struct {
		OwnKind int16  // 拥有者类型 (组织/应用/用户/...)
		OwnID   uint64 // 拥有者ID

		Verify  *LimitVerify  // 验证限制
		Account *LimitAccount // 账号限制
	}

	// LimitVerify 验证限制
	LimitVerify struct {
		BodyLens map[int16]int // [authKind]验证码长度
		Expires  int64         // 过期时间s

		SendDuration int64 // 发送时间范围s
		SendInterval int64 // 发送间隔时间s
		SendMaxTimes int64 // 最大发送次数

		InsertDuration int64 // 添加时间范围s
		InsertInterval int64 // 添加间隔时间s
		InsertMaxTimes int64 // 最大添加次数

		VerifyDuration int64 // 验证时间范围s
		VerifyInterval int64 // 验证间隔时间s
		VerifyMaxTimes int64 // 最大验证次数
	}

	// LimitAuth 认证限制
	LimitAuth struct {
		EnablePhoneCodes  *[]string // []phoneCode 可认证的手机区号，nil不限制
		DisablePhoneCodes *[]string // []phoneCode 禁用的手机区号，nil不限制
		// ...邮件后缀?

		MaxPerUser map[int16]int // [authKind]count 认证最大数量 -1是无限制 一般是1? (user/phone/...)

		EnableScreenPwd bool // 是否启用屏幕密码
	}

	// LimitAccount 账号限制
	LimitAccount struct {
		AuthEnables  []int16 // []authKind 可认证的方式
		AuthRequires []int16 // []authKind 必须认证的方式
		AuthLogins   []int16 // []authKind 可登录的方式 (登录方式里的，不能bind多个账号)

		VerifyRegister   bool // 是否需要验证注册
		VerifyUnRegister bool // 是否需要验证注销

		MaxPerAuthCellphone int  // 单手机可创建的最大数 -1是无限制 0是关闭 一般是1 (pwd只能是1)
		MaxPerAuthEmail     int  // 单邮箱可创建的最大数 -1是无限制 0是关闭 一般是1
		MaxPerAuthBio       int  // 单特征可创建的最大数 -1是无限制 0是关闭 一般是1
		MaxPerAuthThird     int  // 单第三方可创建的最大数 -1是无限制 0是关闭 一般是1
		MaxPerAuthShare     bool // 是否可共享账号最大数量(多个平台最多注册的账号数)
		MaxPerUser          int  // 单用户可创建的最大数 -1是无限制 0是关闭
		MaxPereUserShar     bool // 是否可共享账号最大数量(多个平台最多注册的账号数)

		TokenExpires        int64            // token过期时间 -1是不过期 0是basic 其他是过期时间
		TokenRefreshExpires int64            // refresh过期时间 -1是不过期 0是basic 其他是过期时间
		TokenShares         map[int16]uint64 // [OwnKind]OwnID 可共享token的应用(一般是同Org下的apps)，只用于登录/访问

		NicknameRequire  bool   // 是否需要绑定昵称
		NicknameUnique   bool   // 昵称是否唯一
		NicknameLenRange [2]int // 昵称长度范围

		UserInfoRequire   bool // 是否需要绑定用户信息
		UserBioRequire    bool // 是否需要绑定用户特征
		UserIDCardRequire bool // 是否需要绑定身份证
	}

	LimitAuthPassword struct {
		//MaxPerAcc 只能是1
		MaxPerUser int // 单用户可绑定的最大数
	}
	LimitAuthCellphone struct {
		MaxPerAcc  int // 单账号可绑定的最大数 一般是1
		MaxPerUser int // 单用户可绑定的最大数
	}
	LimitAuthEmail struct {
		MaxPerAcc  int // 单账号可绑定的最大数 一般是1
		MaxPerUser int // 单用户可绑定的最大数
	}
	LimitAuthTelephone struct {
	}

	LimitUser struct {
		//MaxPerAcc 只能是1
		MaxPerCellphone int // 单手机可创建的最大数 -1是无限制 0是关闭 一般是1 (pwd只能是1)
		MaxPerEmail     int // 单邮箱可创建的最大数 -1是无限制 0是关闭 一般是1
		MaxPerBio       int // 单特征可创建的最大数 -1是无限制 0是关闭 一般是1
		MaxPerThird     int // 单第三方可创建的最大数 -1是无限制 0是关闭 一般是1
	}

	//LimitApp struct {
	//}
	//LimitClient struct {
	//}

	LimitDevice struct {
		TrustExpires    int64 // 设备信任 -1是不过期 0是不信任 其他是信任时间
		MaxTrustPerUser int   // 单用户可创建的最大数 -1是无限制 0是关闭
	}
	LimitUserInfo struct {
		RequireSex bool // 是否需要绑定性别
		RequireAge bool // 是否需要绑定年龄
	}
)

func newLimitsDef(ownKind int16, ownID uint64) *Limits {
	return &Limits{
		OwnKind: ownKind,
		OwnID:   ownID,
		Verify:  newLimitVerifyDef(),
	}
}

func newLimitVerifyDef() *LimitVerify {
	return &LimitVerify{
		Expires:        5 * 60,       // 默认过期时间5m
		SendDuration:   1 * 60,       // 默认发送周期1m
		SendInterval:   10,           // 默认发送间隔10s
		SendMaxTimes:   3,            // 默认发送次数3次
		InsertDuration: 12 * 60 * 60, // 默认添加周期12h
		InsertInterval: 60,           // 默认添加间隔60s
		InsertMaxTimes: 10,           // 默认添加次数10次
		VerifyDuration: 1 * 60,       // 默认验证周期1m
		VerifyInterval: 1,            // 默认验证间隔1s
		VerifyMaxTimes: 5,            // 默认验证次数5次
	}
}

func newLimitAccountDef() *LimitAccount {
	return &LimitAccount{
		AuthRequires: []int16{},
		AuthEnables:  []int16{},
		//TokenExpires: make(map[int16]map[uint64]int64),
	}
}

func (s *Base) RegisterLimitVerify(ownKind int16, ownID uint64, limit *LimitVerify) {
	if s.limits == nil {
		s.limits = make(map[int16]map[uint64]*Limits)
	}
	if s.limits[ownKind] == nil {
		s.limits[ownKind] = make(map[uint64]*Limits)
	}
	if s.limits[ownKind][ownID] == nil {
		s.limits[ownKind][ownID] = &Limits{
			OwnKind: ownKind,
			OwnID:   ownID,
			Verify:  limit,
		}
	} else {
		s.limits[ownKind][ownID].Verify = limit
	}
}

func (s *Base) RegisterLimitAccount(ownKind int16, ownID uint64, limit *LimitAccount) {
	if s.limits == nil {
		s.limits = make(map[int16]map[uint64]*Limits)
	}
	if s.limits[ownKind] == nil {
		s.limits[ownKind] = make(map[uint64]*Limits)
	}
	if s.limits[ownKind][ownID] == nil {
		s.limits[ownKind][ownID] = &Limits{
			OwnKind: ownKind,
			OwnID:   ownID,
			Account: limit,
		}
	} else {
		s.limits[ownKind][ownID].Account = limit
	}
}

func (s *Base) GetLimitVerify(ownKind int16, ownID uint64) *LimitVerify {
	if s.limits == nil {
		s.limits = make(map[int16]map[uint64]*Limits)
	}
	if s.limits[ownKind] == nil {
		s.limits[ownKind] = make(map[uint64]*Limits)
	}
	if s.limits[ownKind][ownID] == nil {
		s.limits[ownKind][ownID] = newLimitsDef(ownKind, ownID)
	}
	if s.limits[ownKind][ownID].Verify == nil {
		s.limits[ownKind][ownID].Verify = newLimitVerifyDef()
	}
	return s.limits[ownKind][ownID].Verify
}

func (s *Base) GetLimitAccount(ownKind int16, ownID uint64) *LimitAccount {
	if s.limits == nil {
		s.limits = make(map[int16]map[uint64]*Limits)
	}
	if s.limits[ownKind] == nil {
		s.limits[ownKind] = make(map[uint64]*Limits)
	}
	if s.limits[ownKind][ownID] == nil {
		s.limits[ownKind][ownID] = newLimitsDef(ownKind, ownID)
	}
	if s.limits[ownKind][ownID].Account == nil {
		s.limits[ownKind][ownID].Account = newLimitAccountDef()
	}
	return s.limits[ownKind][ownID].Account
}
