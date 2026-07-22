import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

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

import LineChart from '../LineChart.vue'

describe('LineChart', () => {
  beforeEach(() => (instances.length = 0))

  it('dựng biểu đồ line với nhiều chuỗi', () => {
    mount(LineChart, {
      props: {
        labels: ['2026-07-01', '2026-07-02'],
        series: [
          { label: 'Thu', data: [0, 100], color: '#0a0' },
          { label: 'Chi', data: [5, 0], color: '#a00' },
        ],
      },
    })
    expect(instances).toHaveLength(1)
    expect(instances[0].config.type).toBe('line')
    expect(instances[0].config.data.datasets).toHaveLength(2)
    expect(instances[0].config.data.datasets[0].label).toBe('Thu')
  })

  it('hủy chart khi unmount', () => {
    const w = mount(LineChart, { props: { labels: ['a'], series: [{ label: 'x', data: [1], color: '#1' }] } })
    const inst = instances[instances.length - 1]
    w.unmount()
    expect(inst.destroyed).toBe(true)
  })
})
