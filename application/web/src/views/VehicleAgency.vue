<template>
  <div class="vehicle-agency">
    <div class="app-page-header">
      <a-page-header
        title="车辆管理机构"
        sub-title="负责车辆信息的登记"
        @back="() => $router.push('/')"
      >
        <template #extra>
          <a-tooltip title="点击创建新的车辆信息">
            <a-button type="primary" @click="showCreateModal = true">
              <template #icon><PlusOutlined /></template>
              登记新车辆
            </a-button>
          </a-tooltip>
        </template>
      </a-page-header>
    </div>

    <div class="app-content">
      <a-card :bordered="false">
        <template #extra>
          <div class="card-extra">
            <a-input-search
              v-model:value="searchId"
              placeholder="输入车辆ID进行精确查询"
              style="width: 300px; margin-right: 16px;"
              @search="handleSearch"
              @change="handleSearchChange"
              allow-clear
            />
            <a-radio-group v-model:value="statusFilter" button-style="solid">
              <a-radio-button value="">全部</a-radio-button>
              <a-radio-button value="NORMAL">正常</a-radio-button>
              <a-radio-button value="IN_TRANSACTION">交易中</a-radio-button>
            </a-radio-group>
          </div>
        </template>

        <div class="table-container">
          <a-table
            :columns="columns"
            :data-source="vehicleList"
            :loading="loading"
            :pagination="false"
            :scroll="{ x: 1500, y: 'calc(100vh - 350px)' }"
            row-key="id"
            class="custom-table"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'id'">
                <div class="id-cell">
                  <a-tooltip :title="record.id">
                    <span class="id-text">{{ record.id }}</span>
                  </a-tooltip>
                  <a-tooltip title="点击复制">
                    <copy-outlined
                      class="copy-icon"
                      @click.stop="handleCopy(record.id)"
                    />
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'currentOwner'">
                <div class="id-cell">
                  <a-tooltip :title="record.currentOwner">
                    <span class="id-text">{{ record.currentOwner }}</span>
                  </a-tooltip>
                  <a-tooltip title="点击复制">
                    <copy-outlined
                      class="copy-icon"
                      @click.stop="handleCopy(record.currentOwner)"
                    />
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="record.status === 'NORMAL' ? 'green' : 'blue'">
                  {{ record.status === 'NORMAL' ? '正常' : '交易中' }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'createTime'">
                <time>{{ new Date(record.createTime).toLocaleString() }}</time>
              </template>
              <template v-else-if="column.key === 'updateTime'">
                <time>{{ new Date(record.updateTime).toLocaleString() }}</time>
              </template>
            </template>
          </a-table>
          <div class="load-more">
            <a-button
              :loading="loading"
              @click="loadMore"
              :disabled="!bookmark"
            >
              {{ bookmark ? '加载更多' : '没有更多数据了' }}
            </a-button>
          </div>
        </div>
      </a-card>
    </div>

    <div
      class="block-icon"
      @click="openBlockDrawer"
    >
      <ApartmentOutlined />
    </div>

    <!-- 创建车辆的对话框 -->
    <a-modal
      v-model:visible="showCreateModal"
      title="登记新车辆"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
      :confirmLoading="modalLoading"
    >
      <a-form
        ref="formRef"
        :model="formState"
        :rules="rules"
        layout="vertical"
      >
        <a-form-item label="车型" name="model" extra="请输入车辆型号">
          <a-input-group compact>
            <a-input
              v-model:value="formState.model"
              placeholder="例如: 奔驰 E300L"
              style="width: calc(100% - 110px)"
            />
            <a-tooltip title="随机生成一个车型">
              <a-button @click="generateRandomModelHandler">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <a-form-item label="品牌" name="brand" extra="请输入车辆品牌">
          <a-input-group compact>
            <a-input
              v-model:value="formState.brand"
              placeholder="例如: 奔驰"
              style="width: calc(100% - 110px)"
            />
            <a-tooltip title="随机生成一个品牌">
              <a-button @click="generateRandomBrandHandler">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <a-form-item label="年份" name="year" extra="请输入车辆生产年份">
          <a-input-group compact>
            <a-input-number
              v-model:value="formState.year"
              :min="1990"
              :max="2024"
              style="width: calc(100% - 110px)"
              placeholder="请输入年份"
            />
            <a-tooltip title="随机生成一个年份">
              <a-button @click="generateRandomYearHandler">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <a-form-item label="里程数(公里)" name="mileage" extra="请输入车辆行驶里程">
          <a-input-group compact>
            <a-input-number
              v-model:value="formState.mileage"
              :min="0"
              :step="100"
              style="width: calc(100% - 110px)"
              placeholder="请输入里程数"
            />
            <a-tooltip title="随机生成一个里程数">
              <a-button @click="generateRandomMileageHandler">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <a-form-item label="车况" name="condition" extra="请选择车辆状况">
          <a-select 
            v-model:value="formState.condition"
            placeholder="请选择车况"
          >
            <a-select-option value="优秀">优秀</a-select-option>
            <a-select-option value="良好">良好</a-select-option>
            <a-select-option value="一般">一般</a-select-option>
            <a-select-option value="较差">较差</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="所有者" name="owner" extra="可以输入任意模拟用户名">
          <a-input-group compact>
            <a-input
              v-model:value="formState.owner"
              placeholder="请输入所有者姓名"
              style="width: calc(100% - 110px)"
            />
            <a-tooltip title="随机生成一个用户名">
              <a-button @click="generateRandomOwner">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <div class="form-tips">
          <InfoCircleOutlined style="color: #1890ff; margin-right: 8px;" />
          <span>车辆ID将由系统自动生成</span>
        </div>
      </a-form>
    </a-modal>

    <!-- 区块信息抽屉 -->
    <a-drawer
      v-model:visible="blockDrawer"
      title="区块信息"
      placement="right"
      :width="960"
      :closable="true"
    >
      <div class="block-container">
        <div class="block-header">
          <h3>区块列表</h3>
          <a-pagination
            v-model:current="blockQuery.pageNum"
            v-model:pageSize="blockQuery.pageSize"
            :total="blockTotal"
            :show-total="total => `共 ${total} 条`"
            :page-size-options="['5', '10', '20', '50']"
            show-size-changer
            @change="handleBlockPageChange"
          />
        </div>
        <a-spin :spinning="blockLoading">
          <div class="blocks-wrapper">
            <div 
              v-for="block in blockList" 
              :key="block.blockNum" 
              class="block-item"
            >
              <div class="block-header">
                <h4>区块 #{{ block.blockNum }}</h4>
                <span class="block-time">{{ new Date(block.blockTime).toLocaleString() }}</span>
              </div>
              <div class="block-details">
                <div class="block-detail-item">
                  <span class="label">区块哈希:</span>
                  <div class="value hash">
                    <span class="hash-text">{{ block.blockHash }}</span>
                    <a-tooltip title="点击复制">
                      <copy-outlined
                        class="copy-icon"
                        @click="handleCopy(block.blockHash)"
                      />
                    </a-tooltip>
                  </div>
                </div>
                <div class="block-detail-item">
                  <span class="label">前序哈希:</span>
                  <div class="value hash">
                    <span class="hash-text">{{ block.preBlockHash }}</span>
                    <a-tooltip title="点击复制">
                      <copy-outlined
                        class="copy-icon"
                        @click="handleCopy(block.preBlockHash)"
                      />
                    </a-tooltip>
                  </div>
                </div>
                <div class="block-detail-item">
                  <span class="label">交易数量:</span>
                  <span class="value">{{ block.txCount }}</span>
                </div>
                <div class="block-detail-item">
                  <span class="label">数据哈希:</span>
                  <div class="value hash">
                    <span class="hash-text">{{ block.dataHash }}</span>
                    <a-tooltip title="点击复制">
                      <copy-outlined
                        class="copy-icon"
                        @click="handleCopy(block.dataHash)"
                      />
                    </a-tooltip>
                  </div>
                </div>
              </div>
              <div class="block-txs">
                <h5>区块内交易:</h5>
                <div 
                  v-for="(tx, index) in block.txs" 
                  :key="tx.txId" 
                  class="tx-item"
                >
                  <div class="tx-header">
                    <span class="tx-index">交易 #{{ index + 1 }}</span>
                    <span class="tx-time">{{ new Date(tx.timestamp).toLocaleString() }}</span>
                  </div>
                  <div class="tx-id">
                    <span class="label">交易ID:</span>
                    <div class="value hash">
                      <span class="hash-text">{{ tx.txId }}</span>
                      <a-tooltip title="点击复制">
                        <copy-outlined
                          class="copy-icon"
                          @click="handleCopy(tx.txId)"
                        />
                      </a-tooltip>
                    </div>
                  </div>
                  <div class="tx-creator">
                    <span class="label">创建者:</span>
                    <span class="value creator">{{ tx.creator }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div v-if="blockList.length === 0" class="empty-block">
              <a-empty description="暂无区块数据" />
            </div>
          </div>
        </a-spin>
      </div>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { 
  PlusOutlined, 
  CopyOutlined, 
  InfoCircleOutlined, 
  ReloadOutlined,
  ApartmentOutlined
} from '@ant-design/icons-vue'
import { VehicleAgencyApi } from '@/api'
import { message } from 'ant-design-vue'
import { useClipboard } from '@vueuse/core'
import { generateId, generateRandomName } from '@/utils/random'

const { copy } = useClipboard()

// 数据列表
const vehicleList = ref<any[]>([])
const bookmark = ref<string>('')
const loading = ref<boolean>(false)
const searchId = ref<string>('')
const statusFilter = ref<string>('')

// 表格列定义
const columns = [
  {
    title: '车辆ID',
    dataIndex: 'id',
    key: 'id',
    fixed: 'left',
    width: 220,
  },
  {
    title: '车型',
    dataIndex: 'model',
    key: 'model',
    width: 200,
  },
  {
    title: '品牌',
    dataIndex: 'brand',
    key: 'brand',
    width: 150,
  },
  {
    title: '年份',
    dataIndex: 'year',
    key: 'year',
    width: 100,
  },
  {
    title: '里程数',
    dataIndex: 'mileage',
    key: 'mileage',
    width: 150,
    customRender: ({ text }) => `${text} km`,
  },
  {
    title: '车况',
    dataIndex: 'condition',
    key: 'condition',
    width: 100,
  },
  {
    title: '当前所有者',
    dataIndex: 'currentOwner',
    key: 'currentOwner',
    width: 200,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 200,
  },
  {
    title: '更新时间',
    dataIndex: 'updateTime',
    key: 'updateTime',
    width: 200,
  }
]

// 创建车辆表单
const formRef = ref()
const showCreateModal = ref<boolean>(false)
const modalLoading = ref<boolean>(false)

// 表单数据
const formState = reactive({
  model: '',
  brand: '',
  year: 2020,
  mileage: 0,
  condition: '良好',
  owner: ''
})

// 表单校验规则
const rules = {
  model: [
    { required: true, message: '请输入车型', trigger: 'blur' },
  ],
  brand: [
    { required: true, message: '请输入品牌', trigger: 'blur' },
  ],
  year: [
    { required: true, message: '请输入年份', trigger: 'blur' },
    { type: 'number', min: 1990, message: '年份必须大于等于1990', trigger: 'blur' }
  ],
  mileage: [
    { required: true, message: '请输入里程数', trigger: 'blur' },
    { type: 'number', min: 0, message: '里程数必须大于等于0', trigger: 'blur' }
  ],
  condition: [
    { required: true, message: '请选择车况', trigger: 'change' },
  ],
  owner: [
    { required: true, message: '请输入所有者姓名', trigger: 'blur' },
  ]
}

// 区块信息
const blockDrawer = ref<boolean>(false)
const blockList = ref<any[]>([])
const blockLoading = ref<boolean>(false)
const blockTotal = ref<number>(0)
const blockQuery = reactive({
  pageSize: 5,
  pageNum: 1
})

// 生成随机车型
const generateRandomModelHandler = () => {
  const models = [
    '奔驰 E300L', '宝马 320Li', '奥迪 A4L', '丰田 凯美瑞', '本田 雅阁',
    '大众 帕萨特', '别克 君越', '雷克萨斯 ES', '日产 天籁', '现代 索纳塔',
    '福特 蒙迪欧', '标致 508L', '马自达 阿特兹', '起亚 K5', '沃尔沃 S60'
  ]
  formState.model = models[Math.floor(Math.random() * models.length)]
}

// 生成随机品牌
const generateRandomBrandHandler = () => {
  const brands = [
    '奔驰', '宝马', '奥迪', '丰田', '本田', '大众', '别克', '雷克萨斯',
    '日产', '现代', '福特', '标致', '马自达', '起亚', '沃尔沃'
  ]
  formState.brand = brands[Math.floor(Math.random() * brands.length)]
}

// 生成随机年份
const generateRandomYearHandler = () => {
  formState.year = Math.floor(Math.random() * (2024 - 2000 + 1)) + 2000
}

// 生成随机里程数
const generateRandomMileageHandler = () => {
  formState.mileage = Math.floor(Math.random() * 200000)
}

// 生成随机所有者
const generateRandomOwner = () => {
  formState.owner = generateRandomName()
}

// 查询车辆列表
const fetchVehicleList = async (pageSize = 10, bookmarkVal = '', status = '') => {
  try {
    loading.value = true
    const res = await VehicleAgencyApi.getVehicleList({
      pageSize,
      bookmark: bookmarkVal,
      status
    })
    if (bookmarkVal === '') {
      vehicleList.value = res.records || []
    } else {
      vehicleList.value = [...vehicleList.value, ...(res.records || [])]
    }
    bookmark.value = res.bookmark || ''
  } catch (error) {
    console.error('获取车辆列表失败', error)
    message.error('获取车辆列表失败')
  } finally {
    loading.value = false
  }
}

// 加载更多数据
const loadMore = () => {
  if (bookmark.value) {
    fetchVehicleList(10, bookmark.value, statusFilter.value)
  }
}

// 搜索车辆
const handleSearch = () => {
  if (searchId.value) {
    loading.value = true
    VehicleAgencyApi.getVehicleById(searchId.value).then(res => {
      vehicleList.value = [res]
      bookmark.value = ''
    }).catch(() => {
      message.error('未找到指定车辆')
      vehicleList.value = []
    }).finally(() => {
      loading.value = false
    })
  } else {
    fetchVehicleList(10, '', statusFilter.value)
  }
}

// 搜索输入框变化
const handleSearchChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (!target.value) {
    fetchVehicleList(10, '', statusFilter.value)
  }
}

// 状态筛选变化
watch(statusFilter, () => {
  if (!searchId.value) {
    fetchVehicleList(10, '', statusFilter.value)
  }
})

// 复制文本
const handleCopy = (text: string) => {
  copy(text)
  message.success('复制成功')
}

// 创建车辆对话框确认
const handleModalOk = () => {
  formRef.value.validate().then(() => {
    modalLoading.value = true
    
    VehicleAgencyApi.createVehicle({
      id: generateId(),
      model: formState.model,
      brand: formState.brand,
      year: formState.year,
      mileage: formState.mileage,
      condition: formState.condition,
      owner: formState.owner
    }).then(() => {
      message.success('车辆创建成功')
      showCreateModal.value = false
      // 重置表单
      formRef.value.resetFields()
      // 刷新列表
      fetchVehicleList(10, '', statusFilter.value)
    }).catch(err => {
      message.error(`创建失败: ${err.message}`)
    }).finally(() => {
      modalLoading.value = false
    })
  }).catch(() => {
    // 表单验证失败
  })
}

// 创建车辆对话框取消
const handleModalCancel = () => {
  showCreateModal.value = false
  formRef.value.resetFields()
}

// 查询区块列表
const fetchBlockList = async () => {
  try {
    blockLoading.value = true
    const res = await VehicleAgencyApi.getBlockList({
      pageSize: blockQuery.pageSize,
      pageNum: blockQuery.pageNum
    })
    blockList.value = res.blocks || []
    blockTotal.value = res.total
  } catch (error) {
    console.error('获取区块列表失败', error)
    message.error('获取区块列表失败')
  } finally {
    blockLoading.value = false
  }
}

// 区块分页变化
const handleBlockPageChange = (page: number, pageSize: number) => {
  blockQuery.pageNum = page
  blockQuery.pageSize = pageSize
  fetchBlockList()
}

// 打开区块抽屉
const openBlockDrawer = () => {
  blockDrawer.value = true
  fetchBlockList()
}

// 初始化页面
onMounted(() => {
  fetchVehicleList()
})
</script>

<style scoped>
.vehicle-agency {
  height: 100%;
}
.app-page-header {
  padding: 16px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}
.app-content {
  padding: 16px;
  height: calc(100% - 64px);
}
.card-extra {
  display: flex;
  align-items: center;
}
.table-container {
  margin-top: 16px;
}
.custom-table {
  margin-bottom: 16px;
}
.load-more {
  display: flex;
  justify-content: center;
  margin: 16px 0;
}
.id-cell {
  display: flex;
  align-items: center;
}
.id-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 8px;
}
.copy-icon {
  color: #1890ff;
  cursor: pointer;
}
.form-tips {
  display: flex;
  align-items: center;
  color: rgba(0, 0, 0, 0.45);
  font-size: 14px;
  margin-top: 16px;
}
.block-icon {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 48px;
  height: 48px;
  background: #1890ff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 24px;
  cursor: pointer;
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.16);
  transition: transform 0.3s;
}
.block-icon:hover {
  transform: scale(1.1);
}
.block-container {
  height: 100%;
}
.block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.block-header h3 {
  margin: 0;
}
.blocks-wrapper {
  height: calc(100vh - 160px);
  overflow-y: auto;
  padding-right: 16px;
}
.block-item {
  background: #fafafa;
  border-radius: 4px;
  padding: 16px;
  margin-bottom: 16px;
  border: 1px solid #f0f0f0;
}
.block-item .block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.block-item .block-header h4 {
  margin: 0;
  font-weight: 500;
}
.block-time {
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
}
.block-details {
  margin-bottom: 16px;
}
.block-detail-item {
  margin-bottom: 8px;
  display: flex;
}
.block-detail-item .label {
  width: 80px;
  color: rgba(0, 0, 0, 0.65);
}
.block-detail-item .value {
  flex: 1;
}
.block-detail-item .hash {
  display: flex;
  align-items: center;
}
.hash-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 8px;
  word-break: break-all;
}
.block-txs h5 {
  margin-top: 0;
  margin-bottom: 12px;
  font-weight: 500;
}
.tx-item {
  padding: 12px;
  background: #fff;
  border-radius: 4px;
  margin-bottom: 8px;
  border: 1px solid #f0f0f0;
}
.tx-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}
.tx-index {
  font-weight: 500;
}
.tx-time {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}
.tx-id, .tx-creator {
  display: flex;
  margin-bottom: 4px;
}
.tx-id .label, .tx-creator .label {
  width: 60px;
  color: rgba(0, 0, 0, 0.65);
}
.tx-id .value {
  flex: 1;
  display: flex;
  align-items: center;
}
.creator {
  font-family: monospace;
}
.empty-block {
  padding: 40px 0;
}
</style> 