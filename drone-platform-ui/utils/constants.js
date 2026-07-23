const bizTypeMap = {
  cable_inspection: { label: '电缆巡检', color: '#1976d2' },
  plant_transport: { label: '植保运输', color: '#388e3c' },
  spray_pesticide: { label: '喷洒农药', color: '#f57c00' },
  clean_paint: { label: '清洗喷绘', color: '#7b1fa2' },
  trade_lease: { label: '买卖租赁', color: '#c62828' },
  other: { label: '其他', color: '#616161' }
}

function formatPrice(fen) {
  if (!fen && fen !== 0) return '面议'
  return (fen / 100).toFixed(2)
}

module.exports = { formatPrice, bizTypeMap }
