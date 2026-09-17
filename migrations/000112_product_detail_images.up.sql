-- 商品详情图：详情区可以铺图，不再只有一段描述文字。
--
-- 现状：商品详情区只有 description 一段文字（miniprogram/pkg-eco/pages/mall/detail.vue），
-- 顶部图集（images，≤9 张）放不下的东西——内部结构图、铭牌特写、检测报告、实拍细节——
-- 没有任何地方可放。参考的是 Tigshop 的图文详情思路（descArr），但先做轻量版：
-- 一个独立的详情图数组，不含图文块混排与拖拽排序。
--
-- 与 images 的区别：images 是**顶部图集**（买家第一眼看到的封面），
-- detail_images 是**详情区长图**（决定买家往下翻时能看到多少细节）。两者互不影响。
--
-- 不进审核维度：它是商品内容的一部分，改内容仍走既有的"退回待审核"逻辑
-- （service.UpdateMyProduct），不需要为此单独加字段。

ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS detail_images JSONB NOT NULL DEFAULT '[]'::jsonb;
