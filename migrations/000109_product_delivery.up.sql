-- 交付方式落库。
--
-- 发布表单里早就有「交付方式：自提 / 同城配送 / 物流发货 / 可协商」
-- （miniprogram/utils/publishData.js），但 preview.vue 提交时**根本不发这个字段**——
-- 和"可售数量"一样是采集即丢。结果是"要不要收货地址"只能按商品类型猜：
-- 卖家选了「自提」，买家仍被要求填收货地址。
--
-- 落库后由交付方式决定下单是否必须给地址（service.OrderNeedsReceiver）：
--   logistics / city → 必须给（卖家要送到）
--   pickup           → 不需要（线下自取）
--   negotiable / 未指定 → 按商品类型兜底（实物需要，服务不需要）

ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS delivery TEXT NOT NULL DEFAULT '';

-- 不回填历史数据：空串表示"卖家没选"，一律走商品类型兜底，
-- 与补齐前的行为完全一致（不会因为这次迁移改变任何已有商品的可见性或下单要求）。
