import { mount } from '@vue/test-utils'
import Playground from '@/components/Playground.vue'
import Modal from '@/components/Modal.vue'
import { PrismEditor } from 'vue-prism-editor'
import 'regenerator-runtime'

describe('Playground.vue', () => {
  it('init', async () => {
    const wrapper = mount(Playground, {
      props: { msg: 'message from test', checkWASM: false }
    })
    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toMatch('message from test')
    expect(wrapper.vm.loading).toBe(true)
    expect(wrapper.vm.showAboutModal).toBe(false)
    expect(wrapper.vm.showWASMErrorModal).toBe(false)
    expect(wrapper.vm.result).toBeNull()
    expect(wrapper.vm.code.length).toBeGreaterThan(0)
    const editor = wrapper.getComponent(PrismEditor)
    expect(editor.exists()).toBe(true)
    expect(editor.classes('playground-editor')).toBe(true)
    await wrapper.setData({ code: 'test code' })
    expect(editor.text()).toMatch(/test code$/)
  })

  it('modal', async () => {
    const wrapper = mount(Playground, {
      props: { checkWASM: false }
    })
    expect(wrapper.text()).toMatch(/Run\s+.*\s+About\s+/)
    expect(wrapper.find('div.modal').exists()).toBe(false)
    await wrapper.find('#about-button').trigger('click')
    expect(wrapper.vm.showAboutModal).toBe(true)
    const modal = wrapper.getComponent(Modal)
    expect(modal.exists()).toBe(true)
    expect(modal.vm.showModal).toBe(true)
  })
})
