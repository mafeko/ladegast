import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminLogin from './AdminLogin.vue'
import * as api from '../lib/api'

vi.mock('../lib/api')

describe('AdminLogin', () => {
  it('emits success on correct credentials', async () => {
    vi.mocked(api.adminLogin).mockResolvedValue({ username: 'admin' })
    const wrapper = mount(AdminLogin)

    await wrapper.find('#username').setValue('admin')
    await wrapper.find('#password').setValue('secret')
    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(api.adminLogin).toHaveBeenCalledWith('admin', 'secret')
    expect(wrapper.emitted('success')).toHaveLength(1)
  })

  it('shows an error message on wrong credentials', async () => {
    vi.mocked(api.adminLogin).mockRejectedValue(new Error('invalid credentials'))
    const wrapper = mount(AdminLogin)

    await wrapper.find('form').trigger('submit.prevent')
    await flushMicrotasks()

    expect(wrapper.text()).toContain('invalid credentials')
    expect(wrapper.emitted('success')).toBeUndefined()
  })
})

async function flushMicrotasks() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}
