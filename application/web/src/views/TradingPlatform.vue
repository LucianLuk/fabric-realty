<template>
  <div class="trading-platform"> <!-- 修改类名 -->
    <div class="app-page-header">
      <a-page-header
        title="交易平台"
        sub-title="负责生成交易信息"
        @back="() => $router.push('/')"
      >
        <template #extra>
          <a-tooltip title="点击创建新的交易">
            <a-button type="primary" @click="showCreateModal = true">
              <template #icon><PlusOutlined /></template>
              创建新交易
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
              v-model:value="searchTxId"
              placeholder="输入交易ID进行精确查询"
              style="width: 300px; margin-right: 16px;"
              @search="handleSearchTransaction"
              @change="handleSearchTxChange"
              allow-clear
            />
            <a-radio-group v-model:value="statusFilter" button-style="solid">
              <a-radio-button value="">全部</a-radio-button>
              <a-radio-button value="PENDING">待处理</a-radio-button>
              <a-radio-button value="COMPLETED">已完成</a-radio-button>
              <a-radio-button value="CANCELLED">已取消</a-radio-button>
            </a-radio-group>
          </div>
        </template>

        <div class="table-container">
          <a-table
            :columns="columns"
            :data-source="transactionList"
            :loading="loading"
            :pagination="false"
            :scroll="{ x: 1500, y: 'calc(100vh - 350px)' }"
            row-key="id"
            class="custom-table"
          >
            <template #bodyCell="{ column, record }">
              <!-- ID, Seller, Buyer 复制逻辑保持不变 -->
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
              <!-- 修改为 Car ID -->
              <template v-else-if="column.key === 'carId'">
                <div class="id-cell">
                  <a-tooltip :title="record.carId">
                    <span class="id-text">{{ record.carId }}</span>
                  </a-tooltip>
                  <a-tooltip title="点击复制">
                    <copy-outlined
                      class="copy-icon"
                      @click.stop="handleCopy(record.carId)"
                    />
                  </a-tooltip>
                  <!-- 添加查询汽车详情的链接 -->
                  <a-tooltip title="查询汽车详情">
                    <InfoCircleOutlined
                      class="info-icon"
                      @click.stop="showCarDetails(record.carId)"
                    />
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'seller'">
                <div class="id-cell">
                  <a-tooltip :title="record.seller">
                    <span class="id-text">{{ record.seller }}</span>
                  </a-tooltip>
                  <a-tooltip title="点击复制">
                    <copy-outlined
                      class="copy-icon"
                      @click.stop="handleCopy(record.seller)"
                    />
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'buyer'">
                <div class="id-cell">
                  <a-tooltip :title="record.buyer">
                    <span class="id-text">{{ record.buyer }}</span>
                  </a-tooltip>
                  <a-tooltip title="点击复制">
                    <copy-outlined
                      class="copy-icon"
                      @click.stop="handleCopy(record.buyer)"
                    />
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'price'">
                <span>¥ {{ record.price.toLocaleString() }}</span>
              </template>
              <template v-else-if="column.key === 'status'">
                <a-tag :color="getStatusColor(record.status)">
                  {{ getStatusText(record.status) }}
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

    <!-- 创建交易的对话框 -->
    <a-modal
      v-model:visible="showCreateModal"
      title="创建新交易"
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
        <!-- 修改为 Car ID -->
        <a-form-item label="汽车ID" name="carId" extra="请输入要交易的汽车的ID">
          <a-input-group compact>
            <a-input
              v-model:value="formState.carId"
              placeholder="请输入汽车ID"
              style="width: calc(100% - 110px)"
            />
            <a-tooltip title="随机生成一个ID (仅用于测试)">
              <a-button @click="generateRandomCarId">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <a-form-item label="卖方" name="seller" extra="可以输入任意模拟用户名">
          <a-input-group compact>
            <a-input
              v-model:value="formState.seller"
              placeholder="请输入卖方姓名"
              style="width: calc(100% - 110px)"
            />
            <a-tooltip title="随机生成一个用户名">
              <a-button @click="generateRandomSeller">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <a-form-item label="买方" name="buyer" extra="可以输入任意模拟用户名">
          <a-input-group compact>
            <a-input
              v-model:value="formState.buyer"
              placeholder="请输入买方姓名"
              style="width: calc(100% - 110px)"
            />
            <a-tooltip title="随机生成一个用户名">
              <a-button @click="generateRandomBuyer">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <a-form-item label="价格（元）" name="price" extra="请输入大于0的数值">
          <a-input-group compact>
            <a-input-number
              v-model:value="formState.price"
              :min="0.01"
              :step="0.01"
              style="width: calc(100% - 110px)"
              placeholder="请输入交易价格"
            />
            <a-tooltip title="随机生成一个价格">
              <a-button @click="generateRandomPriceHandler">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
        </a-form-item>

        <div class="form-tips">
          <InfoCircleOutlined style="color: #1890ff; margin-right: 8px;" />
          <span>交易ID将由系统自动生成</span>
        </div>
      </a-form>
    </a-modal>

    <!-- 汽车详情对话框 -->
    <a-modal
      v-model:visible="showCarDetailModal"
      title="汽车详情"
      :footer="null"
      :width="600"
    >
      <a-descriptions v-if="currentCar" bordered :column="1">
        <a-descriptions-item label="汽车ID">{{ currentCar.id }}</a-descriptions-item>
        <a-descriptions-item label="车型">{{ currentCar.model }}</a-descriptions-item>
        <a-descriptions-item label="VIN">{{ currentCar.vin }}</a-descriptions-item>
        <a-descriptions-item label="当前所有者">{{ currentCar.currentOwner }}</a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="currentCar.status === 'AVAILABLE' ? 'green' : (currentCar.status === 'IN_TRANSACTION' ? 'blue' : 'orange')">
            {{ currentCar.status === 'AVAILABLE' ? '待售' : (currentCar.status === 'IN_TRANSACTION' ? '交易中' : '已售') }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">{{ new Date(currentCar.createTime).toLocaleString() }}</a-descriptions-item>
        <a-descriptions-item label="更新时间">{{ new Date(currentCar.updateTime).toLocaleString() }}</a-descriptions-item>
      </a-descriptions>
      <a-spin v-else />
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

        <div class="block-list">
          <a-card v-for="block in blockList" :key="block.block_num" class="block-item">
            <template #title>
              <div class="block-item-header">
                <span class="block-number">区块 #{{ block.block_num }}</span>
                <span class="block-time">{{ new Date(block.save_time).toLocaleString() }}</span>
              </div>
            </template>
            <div class="block-item-content">
              <div class="block-field">
                <span class="field-label">区块哈希：</span>
                <a-tooltip :title="block.block_hash">
                  <span class="field-value hash">{{ block.block_hash }}</span>
                </a-tooltip>
                <a-tooltip title="点击复制">
                  <copy-outlined
                    class="copy-icon"
                    @click="handleCopy(block.block_hash)"
                  />
                </a-tooltip>
              </div>
              <div class="block-field">
                <span class="field-label">数据哈希：</span>
                <a-tooltip :title="block.data_hash">
                  <span class="field-value hash">{{ block.data_hash }}</span>
                </a-tooltip>
                <a-tooltip title="点击复制">
                  <copy-outlined
                    class="copy-icon"
                    @click="handleCopy(block.data_hash)"
                  />
                </a-tooltip>
              </div>
              <div class="block-field">
                <span class="field-label">前块哈希：</span>
                <a-tooltip :title="block.prev_hash">
                  <span class="field-value hash">{{ block.prev_hash }}</span>
                </a-tooltip>
                <a-tooltip title="点击复制">
                  <copy-outlined
                    class="copy-icon"
                    @click="handleCopy(block.prev_hash)"
                  />
                </a-tooltip>
              </div>
              <div class="block-field">
                <span class="field-label">交易数量：</span>
                <span class="field-value">{{ block.tx_count }}</span>
              </div>
            </div>
          </a-card>
        </div>
      </div>
    </a-drawer>

  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue';
import { PlusOutlined, InfoCircleOutlined, ReloadOutlined, CopyOutlined, ApartmentOutlined } from '@ant-design/icons-vue';
import { tradingPlatformApi } from '../api'; // API 导入保持不变
import type { FormInstance } from 'ant-design-vue';
import { ref, reactive, watch, onMounted } from 'vue';
import type { BlockData, Transaction, Car } from '../types'; // 导入 Car 类型
import { copyToClipboard, generateRandomName, generateRandomPrice, generateUUID } from '../utils';

const formRef = ref<FormInstance>();
const showCreateModal = ref(false);
const modalLoading = ref(false);

// 修改表单状态
const formState = reactive({
  carId: '', // 修改为 carId
  seller: '',
  buyer: '',
  price: undefined as number | undefined,
});

// 修改验证规则
const rules = {
  carId: [{ required: true, message: '请输入汽车ID' }], // 修改为 carId
  seller: [{ required: true, message: '请输入卖方' }],
  buyer: [{ required: true, message: '请输入买方' }],
  price: [
    { required: true, message: '请输入价格' },
    { type: 'number', min: 0.01, message: '价格必须大于0' }
  ],
};

// 修改列定义
const columns = [
  {
    title: '交易ID',
    dataIndex: 'id',
    key: 'id',
    width: 180,
    ellipsis: false,
    customCell: () => ({ style: { whiteSpace: 'nowrap', overflow: 'hidden' } }),
  },
  {
    title: '汽车ID', // 修改标题
    dataIndex: 'carId', // 修改 dataIndex
    key: 'carId', // 修改 key
    width: 180,
    ellipsis: false,
    customCell: () => ({ style: { whiteSpace: 'nowrap', overflow: 'hidden' } }),
  },
  {
    title: '卖方',
    dataIndex: 'seller',
    key: 'seller',
    width: 120,
    ellipsis: false,
    customCell: () => ({ style: { whiteSpace: 'nowrap', overflow: 'hidden' } }),
  },
  {
    title: '买方',
    dataIndex: 'buyer',
    key: 'buyer',
    width: 120,
    ellipsis: false,
    customCell: () => ({ style: { whiteSpace: 'nowrap', overflow: 'hidden' } }),
  },
  {
    title: '价格 (元)',
    dataIndex: 'price',
    key: 'price',
    width: 120,
    align: 'right',
    customCell: () => ({ style: { fontVariantNumeric: 'tabular-nums' } }),
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
    width: 180,
  },
  {
    title: '更新时间',
    dataIndex: 'updateTime',
    key: 'updateTime',
    width: 180,
  },
];

const transactionList = ref<Transaction[]>([]); // 类型保持 Transaction
const loading = ref(false);
const bookmark = ref('');

// 加载交易列表函数 (API 调用不变)
const loadTransactionList = async () => {
  try {
    loading.value = true;
    const result = await tradingPlatformApi.getTransactionList({
      pageSize: 10,
      bookmark: bookmark.value,
      status: statusFilter.value,
    });
    if (!bookmark.value) {
      transactionList.value = result.records;
    } else {
      transactionList.value = [...transactionList.value, ...result.records];
    }
    bookmark.value = result.bookmark;
  } catch (error: any) {
    message.error(error.message || '加载交易列表失败');
  } finally {
    loading.value = false;
  }
};

const loadMore = () => {
  loadTransactionList();
};

// 修改模态框确认逻辑
const handleModalOk = () => {
  formRef.value?.validate().then(async () => {
    modalLoading.value = true;
    try {
      const transactionData = { // 修改变量名
        ...formState,
        txId: generateUUID(), // 生成交易ID
      };
      await tradingPlatformApi.createTransaction(transactionData); // API 调用不变，但参数已修改
      message.success('交易创建成功');
      showCreateModal.value = false;
      formRef.value?.resetFields();
      transactionList.value = [];
      bookmark.value = '';
      loadTransactionList();
    } catch (error: any) {
      message.error(error.message || '交易创建失败');
    } finally {
      modalLoading.value = false;
    }
  });
};

const handleModalCancel = () => {
  showCreateModal.value = false;
  formRef.value?.resetFields();
};

// 修改随机生成函数
const generateRandomCarId = () => {
  formState.carId = generateUUID().substring(0, 8); // 简单生成一个随机ID
};
const generateRandomSeller = () => {
  formState.seller = generateRandomName();
};
const generateRandomBuyer = () => {
  formState.buyer = generateRandomName();
};
const generateRandomPriceHandler = () => {
  formState.price = generateRandomPrice();
};

const handleCopy = (text: string) => {
  copyToClipboard(text);
};

const statusFilter = ref('');

watch(statusFilter, () => {
  transactionList.value = [];
  bookmark.value = '';
  loadTransactionList();
});

const searchTxId = ref('');

// 搜索交易逻辑 (API 调用不变)
const handleSearchTransaction = async (value: string) => {
  if (!value) {
    message.warning('请输入要查询的交易ID');
    return;
  }
  try {
    const result = await tradingPlatformApi.getTransaction(value);
    transactionList.value = [result];
    bookmark.value = '';
  } catch (error: any) {
    message.error(error.message || '查询交易信息失败');
    transactionList.value = [];
  }
};

const handleSearchTxChange = (e: Event) => {
  const value = (e.target as HTMLInputElement).value;
  if (!value) {
    transactionList.value = [];
    bookmark.value = '';
    loadTransactionList();
  }
};

// 状态显示逻辑
const getStatusColor = (status: Transaction['status']) => {
  switch (status) {
    case 'PENDING': return 'processing';
    case 'COMPLETED': return 'success';
    case 'CANCELLED': return 'error';
    default: return 'default';
  }
};
const getStatusText = (status: Transaction['status']) => {
  switch (status) {
    case 'PENDING': return '待处理';
    case 'COMPLETED': return '已完成';
    case 'CANCELLED': return '已取消';
    default: return '未知';
  }
};

onMounted(() => {
  loadTransactionList();
});

// 区块信息部分 (API 调用不变)
const blockDrawer = ref(false);
const blockList = ref<BlockData[]>([]);
const blockTotal = ref(0);
const blockQuery = reactive({
  pageSize: 10,
  pageNum: 1,
});

const openBlockDrawer = async () => {
  blockDrawer.value = true;
  await fetchBlockList();
};

const fetchBlockList = async () => {
  try {
    const res = await tradingPlatformApi.getBlockList({
      pageSize: blockQuery.pageSize,
      pageNum: blockQuery.pageNum,
    });
    blockList.value = res.blocks;
    blockTotal.value = res.total;
  } catch (error) {
    console.error('获取区块列表失败：', error);
  }
};

const handleBlockPageChange = async (page: number, pageSize: number) => {
  blockQuery.pageNum = page;
  blockQuery.pageSize = pageSize;
  await fetchBlockList();
};

// 新增：汽车详情模态框逻辑
const showCarDetailModal = ref(false);
const currentCar = ref<Car | null>(null);

const showCarDetails = async (carId: string) => {
  currentCar.value = null; // 重置
  showCarDetailModal.value = true;
  try {
    // 使用交易平台的 API 查询汽车信息
    currentCar.value = await tradingPlatformApi.getCar(carId);
  } catch (error: any) {
    message.error(error.message || '查询汽车详情失败');
    showCarDetailModal.value = false; // 查询失败则关闭模态框
  }
};

</script>

<style scoped>
/* 样式可以保持不变，或者根据需要调整 */
.trading-platform { /* 保持与 template 中一致 */
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f0f2f5;
}

.app-page-header {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.app-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.card-extra {
  display: flex;
  align-items: center;
}

.table-container {
  margin-top: 16px;
}

.custom-table .id-cell {
  display: flex;
  align-items: center;
}

.custom-table .id-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px; /* 根据需要调整 */
  display: inline-block;
  vertical-align: middle;
}

.custom-table .copy-icon {
  margin-left: 8px;
  cursor: pointer;
  color: #1890ff;
  vertical-align: middle;
}

/* 新增：汽车详情图标样式 */
.custom-table .info-icon {
  margin-left: 8px;
  cursor: pointer;
  color: #1890ff;
  vertical-align: middle;
}


.load-more {
  text-align: center;
  margin-top: 16px;
}

.block-icon {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background-color: #1890ff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.3s;
}

.block-icon:hover {
  background-color: #40a9ff;
}

.form-tips {
  color: rgba(0, 0, 0, 0.45);
  font-size: 14px;
  margin-top: 8px;
  display: flex;
  align-items: center;
}

.block-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.block-list {
  flex: 1;
  overflow-y: auto;
  padding-right: 8px; /* 防止滚动条遮挡 */
}

.block-item {
  margin-bottom: 16px;
}

.block-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.block-number {
  font-weight: 500;
}

.block-time {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}

.block-item-content {
  font-size: 13px;
}

.block-field {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
}

.field-label {
  font-weight: 500;
  color: rgba(0, 0, 0, 0.65);
  margin-right: 8px;
  min-width: 70px; /* 调整对齐 */
}

.field-value {
  flex: 1;
  word-break: break-all;
}

.field-value.hash {
  font-family: 'Courier New', Courier, monospace;
  max-width: calc(100% - 100px); /* 调整防止过长 */
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  vertical-align: middle;
}

.block-item-content .copy-icon {
  margin-left: 8px;
  cursor: pointer;
  color: #1890ff;
  vertical-align: middle;
}

</style>
