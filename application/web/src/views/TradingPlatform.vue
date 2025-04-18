<template>
  <div class="trading-platform">
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
              生成新交易
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
              placeholder="输入交易ID进行精确查询"
              style="width: 300px; margin-right: 16px;"
              @search="handleSearch"
              @change="handleSearchChange"
              allow-clear
            />
            <a-radio-group v-model:value="statusFilter" button-style="solid">
              <a-radio-button value="">全部</a-radio-button>
              <a-radio-button value="PENDING">待完成</a-radio-button>
              <a-radio-button value="COMPLETED">已完成</a-radio-button>
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
              <template v-else-if="column.key === 'vehicleId'">
                <div class="id-cell">
                  <a-tooltip :title="record.vehicleId">
                    <span class="id-text">{{ record.vehicleId }}</span>
                  </a-tooltip>
                  <a-tooltip title="点击复制">
                    <copy-outlined
                      class="copy-icon"
                      @click.stop="handleCopy(record.vehicleId)"
                    />
                  </a-tooltip>
                </div>
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

    <!-- 生成交易的对话框 -->
    <a-modal
      v-model:visible="showCreateModal"
      title="生成新交易"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
      :confirmLoading="modalLoading"
      :style="{ top: '40px' }"
    >
      <a-form
        ref="formRef"
        :model="formState"
        :rules="rules"
        layout="vertical"
      >
        <a-form-item label="车辆ID" name="vehicleId" extra="请输入要交易的车辆ID">
          <a-input
            v-model:value="formState.vehicleId"
            placeholder="请输入车辆ID"
            @change="handleVehicleIdChange"
          />
        </a-form-item>

        <a-form-item label="卖家" name="seller" extra="当前车辆所有者">
          <a-input
            v-model:value="formState.seller"
            placeholder="自动填入当前所有者"
            disabled
          />
        </a-form-item>

        <a-form-item label="买家" name="buyer" extra="可以输入任意模拟用户名作为买家">
          <a-input-group compact>
            <a-input
              v-model:value="formState.buyer"
              placeholder="请输入买家姓名"
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

        <a-form-item label="价格" name="price" extra="请输入大于0的交易金额">
          <a-input-group compact>
            <a-input-number
              v-model:value="formState.price"
              :min="0.01"
              :step="0.01"
              style="width: calc(100% - 110px)"
              placeholder="请输入价格"
              :formatter="value => `¥ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
              :parser="value => value!.replace(/\¥\s?|(,*)/g, '')"
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

    <div
      class="block-icon"
      @click="openBlockDrawer"
    >
      <ApartmentOutlined />
    </div>

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
import { PlusOutlined, InfoCircleOutlined, CopyOutlined, ReloadOutlined, ApartmentOutlined } from '@ant-design/icons-vue';
import { tradingPlatformApi } from '../api';
import type { FormInstance } from 'ant-design-vue';
import { ref, reactive } from 'vue';
import type { BlockData } from '../types';
import { copyToClipboard, generateRandomName, generateRandomPrice, generateUUID, getStatusText, getStatusColor, formatPrice } from '../utils';

const formRef = ref<FormInstance>();
const showCreateModal = ref(false);
const modalLoading = ref(false);

const formState = reactive({
  vehicleId: '',
  seller: '',
  buyer: '',
  price: undefined as number | undefined,
});

const rules = {
  vehicleId: [{ required: true, message: '请输入车辆ID' }],
  seller: [{ required: true, message: '请输入卖家' }],
  buyer: [{ required: true, message: '请输入买家' }],
  price: [
    { required: true, message: '请输入价格' },
    { type: 'number', min: 0.01, message: '价格必须大于0' }
  ],
};

const columns = [
  {
    title: '交易ID',
    dataIndex: 'id',
    key: 'id',
    width: 180,
    ellipsis: false,
    customCell: () => ({
      style: {
        whiteSpace: 'nowrap',
        overflow: 'hidden',
      }
    }),
  },
  {
    title: '车辆ID',
    dataIndex: 'vehicleId',
    key: 'vehicleId',
    width: 180,
    ellipsis: false,
    customCell: () => ({
      style: {
        whiteSpace: 'nowrap',
        overflow: 'hidden',
      }
    }),
  },
  {
    title: '卖家',
    dataIndex: 'seller',
    key: 'seller',
    width: 120,
    ellipsis: true,
  },
  {
    title: '买家',
    dataIndex: 'buyer',
    key: 'buyer',
    width: 120,
    ellipsis: true,
  },
  {
    title: '价格',
    dataIndex: 'price',
    key: 'price',
    width: 120,
    customRender: ({ text }: { text: number }) => formatPrice(text),
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

const transactionList = ref<any[]>([]);
const loading = ref(false);
const bookmark = ref('');

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

const getStatusColor = (status: string) => {
  switch (status) {
    case 'PENDING':
      return 'blue';
    case 'COMPLETED':
      return 'green';
    default:
      return 'default';
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case 'PENDING':
      return '待完成';
    case 'COMPLETED':
      return '已完成';
    default:
      return '未知';
  }
};

const handleModalOk = () => {
  formRef.value?.validate().then(async () => {
    modalLoading.value = true;
    try {
      const transactionData = {
        ...formState,
        txId: generateUUID(),
      };
      await tradingPlatformApi.createTransaction(transactionData);
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

const statusFilter = ref('');

watch(statusFilter, () => {
  transactionList.value = [];
  bookmark.value = '';
  loadTransactionList();
});

const handleCopy = (text: string) => {
  copyToClipboard(text);
};

const handleVehicleIdChange = async (e: Event) => {
  const id = (e.target as HTMLInputElement).value;
  if (!id) {
    formState.seller = '';
    return;
  }

  try {
    const result = await tradingPlatformApi.getVehicle(id);
    if (result.status !== 'NORMAL') {
      message.error('该车辆不是正常状态，无法生成交易');
      formState.vehicleId = '';
      formState.seller = '';
      return;
    }
    formState.seller = result.currentOwner;
  } catch (error: any) {
    message.error(error.message || '获取车辆信息失败');
    formState.seller = '';
  }
};

const generateRandomBuyer = () => {
  formState.buyer = generateRandomName();
};

const generateRandomPriceHandler = () => {
  formState.price = generateRandomPrice();
};

const searchId = ref('');

const handleSearch = async (value: string) => {
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

const handleSearchChange = (e: Event) => {
  const value = (e.target as HTMLInputElement).value;
  if (!value) {
    transactionList.value = [];
    bookmark.value = '';
    loadTransactionList();
  }
};

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

onMounted(() => {
  loadTransactionList();
});
</script>

<style scoped>
</style>
