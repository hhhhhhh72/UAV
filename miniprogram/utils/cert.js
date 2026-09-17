// 认证判定（单一来源）。
//
// 这三个函数原本是 pages/demands/detail.vue 的页内函数，只有接单流程在用。
// 商品发布现在也要查企业认证（商品详情页对买家承诺「平台认证商家」，发布必须同等门槛），
// 再抄一份必然与这里漂移，所以抽到 utils 共用。
import { request } from './request'

// 企业认证：是否存在 status=approved 的企业记录
export async function isEnterpriseCertified() {
  try {
    const res = await request({ url: '/api/v1/enterprises' })
    const data = (res && res.data) || res || {}
    const items = Array.isArray(data) ? data : (data && data.items) || []
    return items.some((e) => e && e.status === 'approved')
  } catch (e) {
    return false
  }
}

// 飞手认证：个人飞手走此通道（/api/v1/certified-pilots/mine）
export async function isPilotCertified() {
  try {
    const res = await request({ url: '/api/v1/certified-pilots/mine' })
    const p = res || {}
    return p.status === 'approved'
  } catch (e) {
    return false
  }
}

// 接单认证门槛：企业认证或飞手认证任一通过即可申请接单（个人飞手不强制企业主体）
export async function isAnyCertified() {
  return (await isEnterpriseCertified()) || (await isPilotCertified())
}
