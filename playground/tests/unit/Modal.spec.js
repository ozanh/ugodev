import 'regenerator-runtime'

import { mount } from '@vue/test-utils'
import Modal from '@/components/Modal.vue'

describe('Modal.vue', () => {
  it('showModal false', () => {
    const wrapper = mount(Modal, {
      props: { showModal: false }
    })
    expect(wrapper.text()).toBe('')
    expect(wrapper.vm.showModal).toBe(false)
  })

  it('showModal true slots', async () => {
    const wrapper = mount(Modal, {
      props: { showModal: true },
      slots: {
        body: 'body-holder',
        footer: 'footer-holder'
      }
    })
    expect(wrapper.text()).toBe('body-holderfooter-holder')
    expect(wrapper.vm.showModal).toBe(true)
    await wrapper.setProps({ showModal: false })
    expect(wrapper.vm.showModal).toBe(false)
    expect(wrapper.text()).toBe('')
  })
})
