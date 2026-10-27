<template>
  <div class="relative">
    <input
      :id="id || undefined"
      :value="modelValue"
      :type="inputType"
      :placeholder="placeholder"
      :disabled="disabled"
      :required="required"
      :minlength="minlength"
      :autocomplete="autocomplete || (asPassword ? 'new-password' : 'off')"
      :name="name || (asPassword ? undefined : `field-${uid}`)"
      autocapitalize="off"
      autocorrect="off"
      spellcheck="false"
      data-lpignore="true"
      data-1p-ignore="true"
      data-form-type="other"
      class="peer input-field pr-10"
      :class="{ 'input-field-icon': icon, 'secret-masked': !asPassword && !visible }"
      @input="$emit('update:modelValue', $event.target.value)"
    />
    <Icon
      v-if="icon"
      :name="icon"
      class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400 transition-colors peer-focus:text-blue-500"
    />
    <button
      type="button"
      tabindex="-1"
      class="absolute inset-y-0 right-0 px-3 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
      :title="visible ? '隐藏' : '显示'"
      :aria-label="visible ? '隐藏' : '显示'"
      @click="visible = !visible"
    >
      <Icon :name="visible ? 'eye-off' : 'eye'" class="h-4 w-4" />
    </button>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import Icon from './Icon.vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  required: { type: Boolean, default: false },
  minlength: { type: Number, default: undefined },
  asPassword: { type: Boolean, default: false },
  // Optional leading icon (an Icon name).
  icon: { type: String, default: '' },
  id: { type: String, default: '' },
  // Login forms pass name/autocomplete so password managers still work.
  name: { type: String, default: '' },
  autocomplete: { type: String, default: '' }
})

defineEmits(['update:modelValue'])

const visible = ref(false)
const uid = Math.random().toString(36).slice(2, 8)

const inputType = computed(() => {
  if (props.asPassword) {
    return visible.value ? 'text' : 'password'
  }
  return 'text'
})
</script>
