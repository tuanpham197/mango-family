import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ProfileView from '../ProfileView.vue'
import { useAuthStore } from '../../stores/auth'

const push = vi.fn()
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))

describe('ProfileView', () => {
  it('hiển thị tên, email, hộ + avatar viết tắt', () => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    auth.user = { id: 'u1', email: 'alice@dev.local', display_name: 'Alice Nguyen' }
    auth.household = { id: 'h1', name: 'Hộ A' }

    const w = mount(ProfileView)
    expect(w.get('[data-testid="profile-name"]').text()).toBe('Alice Nguyen')
    expect(w.get('[data-testid="profile-email"]').text()).toBe('alice@dev.local')
    expect(w.get('[data-testid="profile-household"]').text()).toBe('Hộ A')
    expect(w.get('[data-testid="profile-avatar"]').text()).toBe('AN')
  })

  it('nút Quay lại → về trang chủ', async () => {
    setActivePinia(createPinia())
    const w = mount(ProfileView)
    await w.get('[data-testid="profile-back"]').trigger('click')
    expect(push).toHaveBeenCalledWith('/')
  })
})
