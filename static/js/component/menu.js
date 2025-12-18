import { ref } from 'vue'

export default {
    setup(){
        const count = ref(0)
        const increment = () => {
            count.value += 1
        }
        return {count, increment}
    },
    template: `
    <a-tooltip placement="topLeft">
        <template #title>
            <span>计数器</span>
        </template>
        <a-button type="primary">
            <router-link to="/counter">
                <span>计数器</span>
                <span class="badge">{{ count }}</span>
                </router-link>
        </a-button>
    </a-tooltip>
    `,
}
