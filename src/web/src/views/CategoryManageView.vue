<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ApiError } from '../api/client'
import type { Category, CategoryType } from '../api/types'
import DeleteReassignDialog from '../components/DeleteReassignDialog.vue'
import { useCategoriesStore } from '../stores/categories'
import { useInvalidation } from '../composables/useInvalidation'

const categories = useCategoriesStore()
const tab = ref<CategoryType>('EXPENSE')
const deleting = ref<Category | null>(null)
const notice = ref('')

onMounted(() => categories.ensure())
// Thành viên khác sửa danh mục → danh sách tự làm mới ≤ 5s (D8, quickstart #13).
useInvalidation('categories_changed', () => categories.fetch())

const groups = computed(() => categories.byType(tab.value))

async function toggleHidden(cat: Category) {
  notice.value = ''
  try {
    await categories.update(cat.id, {
      is_hidden: !cat.is_hidden,
      expected_updated_at: cat.updated_at,
    })
  } catch (e) {
    if (e instanceof ApiError) {
      notice.value = e.message
      await categories.fetch() // conflict/gone → đồng bộ lại trạng thái mới nhất
    }
  }
}

function doneDelete(message?: string) {
  deleting.value = null
  if (message) notice.value = message
}
</script>

<template>
  <div>
    <div class="page-head">
      <h1>Danh mục</h1>
      <RouterLink class="btn btn-primary" to="/categories/new" data-testid="add-category">＋ Thêm</RouterLink>
    </div>

    <div class="segments" data-testid="category-tabs">
      <button type="button" :class="{ 'active-expense': tab === 'EXPENSE' }" data-testid="tab-expense" @click="tab = 'EXPENSE'">
        Chi phí
      </button>
      <button type="button" :class="{ 'active-income': tab === 'INCOME' }" data-testid="tab-income" @click="tab = 'INCOME'">
        Thu nhập
      </button>
    </div>

    <div v-if="notice" class="form-error" data-testid="notice">{{ notice }}</div>

    <div class="card">
      <p v-if="groups.length === 0" class="muted">Chưa có danh mục nào.</p>
      <template v-for="root in groups" :key="root.id">
        <div class="list-row" data-testid="category-row">
          <span class="row-icon">{{ root.icon ?? '🏷️' }}</span>
          <div class="row-main">
            <div class="row-title">
              {{ root.name }}<span v-if="root.is_hidden" class="badge-hidden">Đã ẩn</span>
            </div>
          </div>
          <div class="row-actions">
            <RouterLink class="action" :to="`/categories/${root.id}/edit`" :data-testid="`edit-${root.name}`">Sửa</RouterLink>
            <button type="button" class="action" :data-testid="`toggle-hidden-${root.name}`" @click="toggleHidden(root)">
              {{ root.is_hidden ? 'Bỏ ẩn' : 'Ẩn' }}
            </button>
            <button type="button" class="action danger" :data-testid="`delete-${root.name}`" @click="deleting = root">
              Xóa
            </button>
          </div>
        </div>
        <div v-for="child in root.children" :key="child.id" class="list-row child-row" data-testid="category-row">
          <span class="row-icon">{{ child.icon ?? '🏷️' }}</span>
          <div class="row-main">
            <div class="row-title">
              {{ child.name }}<span v-if="child.is_hidden" class="badge-hidden">Đã ẩn</span>
            </div>
            <div class="row-sub">thuộc {{ root.name }}</div>
          </div>
          <div class="row-actions">
            <RouterLink class="action" :to="`/categories/${child.id}/edit`" :data-testid="`edit-${child.name}`">Sửa</RouterLink>
            <button type="button" class="action" :data-testid="`toggle-hidden-${child.name}`" @click="toggleHidden(child)">
              {{ child.is_hidden ? 'Bỏ ẩn' : 'Ẩn' }}
            </button>
            <button type="button" class="action danger" :data-testid="`delete-${child.name}`" @click="deleting = child">
              Xóa
            </button>
          </div>
        </div>
      </template>
    </div>

    <DeleteReassignDialog v-if="deleting" :category="deleting" @close="deleting = null" @done="doneDelete" />
  </div>
</template>

<style scoped>
.child-row {
  padding-left: 24px;
}

.row-actions {
  display: flex;
  gap: 4px;
}

.action {
  border: none;
  background: transparent;
  color: var(--green);
  font: inherit;
  font-size: 13px;
  cursor: pointer;
  text-decoration: none;
  padding: 4px 6px;
}

.action.danger {
  color: var(--red);
}
</style>
