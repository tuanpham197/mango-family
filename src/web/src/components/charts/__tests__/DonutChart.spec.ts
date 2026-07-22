import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

// Mock chart.js registry: FakeChart ghi lại config + trạng thái destroy.
const { instances, FakeChart } = vi.hoisted(() => {
  const instances: Array<{ config: any; destroyed: boolean }> = []
  class FakeChart {
    config: any
    data: any
    destroyed = false
    constructor(_canvas: unknown, config: any) {
      this.config = config
      this.data = config.data
      instances.push(this)
    }
    update() {}
    destroy() {
      this.destroyed = true
    }
  }
  return { instances, FakeChart }
})
vi.mock('../registry', () => ({ Chart: FakeChart }))

import DonutChart from '../DonutChart.vue'

describe('DonutChart', () => {
  beforeEach(() => (instances.length = 0))

  it('dựng biểu đồ doughnut với nhãn/giá trị/màu', () => {
    mount(DonutChart, { props: { labels: ['Ăn uống', 'Di chuyển'], values: [3500000, 900000], colors: ['#1', '#2'] } })
    expect(instances).toHaveLength(1)
    expect(instances[0].config.type).toBe('doughnut')
    expect(instances[0].config.data.datasets[0].data).toEqual([3500000, 900000])
  })

  it('hủy chart khi unmount', () => {
    const w = mount(DonutChart, { props: { labels: ['A'], values: [1], colors: ['#1'] } })
    const inst = instances[instances.length - 1]
    w.unmount()
    expect(inst.destroyed).toBe(true)
  })
})
