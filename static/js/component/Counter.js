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
   <div class="counter-container">
      <a-card title="计数器 Demo">
        <a-statistic :value="count"/>
        <a-button 
          type="primary" 
          @click="increment"
          class="mt-2">
          + 增加
        </a-button>
      </a-card>
    </div>
    `,
}
