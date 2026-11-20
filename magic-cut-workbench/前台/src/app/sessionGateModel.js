export function createSessionGateModel(session, locationLike) {
  const unavailable = session?.errorCode === 'ACCOUNT_SERVICE_UNAVAILABLE'
  return {
    title: unavailable ? '无法连接一战晟铭账号服务' : '请从一战晟铭进入魔剪',
    message: unavailable
      ? '暂时无法读取当前账号，请返回一战晟铭确认服务正常后重新进入魔剪。'
      : '当前魔剪会话已失效，请返回一战晟铭后重新切换到魔剪。',
    returnLabel: '返回一战晟铭',
    returnHref: `${locationLike.protocol}//${locationLike.hostname}:4001/`,
  }
}
