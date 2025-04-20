<template>
  <div class="car-dealer">
    <div class="app-page-header">
      <a-page-header
        title="汽车经销商"
        sub-title="负责汽车信息的登记"
        @back="() => $router.push('/')"
      >
        <template #extra>
          <a-tooltip title="点击创建新的汽车信息">
            <a-button type="primary" @click="showCreateModal = true">
              <template #icon><PlusOutlined /></template>
              登记新汽车
            </a-button>
          </a-tooltip>
        </template>
      </a-page-header>
    </div>

    <div class="app-content">
      <a-card :bordered="false">
        <!-- Removed the entire <template #extra> slot as a workaround -->
        <!-- <template #extra>
          <div class="card-extra">
             <a-radio-group v-model:value="statusFilter" button-style="solid" style="margin-right: 16px;">
              <a-radio-button value="">全部</a-radio-button>
              <a-radio-button value="AVAILABLE">待售</a-radio-button>
              <a-radio-button value="IN_TRANSACTION">交易中</a-radio-button>
              <a-radio-button value="SOLD">已售</a-radio-button>
            </a-radio-group>
            <a-input-search
              v-model:value="searchId"
              placeholder="输入汽车ID进行精确查询"
              style="width: 300px;"
              @search="handleSearch"
              @change="handleSearchChange"
              allow-clear
            ></a-input-search>
          </div>
        </template> -->

        <div class="table-container">
          <a-table
            :columns="columns"
            :data-source="carList"
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
              <template v-else-if="column.key === 'model'">
                 <div class="id-cell">
                  <a-tooltip :title="record.model">
                    <span class="id-text">{{ record.model }}</span>
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'vin'">
                 <div class="id-cell">
                  <a-tooltip :title="record.vin">
                    <span class="id-text">{{ record.vin }}</span>
                  </a-tooltip>
                  <a-tooltip title="点击复制">
                    <copy-outlined
                      class="copy-icon"
                      @click.stop="handleCopy(record.vin)"
                    />
                  </a-tooltip>
                </div>
              </template>
              <template v-else-if="column.key === 'status'">
                 <a-tag :color="record.status === 'AVAILABLE' ? 'green' : (record.status === 'IN_TRANSACTION' ? 'blue' : 'orange')">
                  {{ record.status === 'AVAILABLE' ? '待售' : (record.status === 'IN_TRANSACTION' ? '交易中' : '已售') }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'createTime'">
                <time>{{ new Date(record.createTime).toLocaleString() }}</time>
              </template>
              <template v-else-if="column.key === 'updateTime'">
                <time>{{ new Date(record.updateTime).toLocaleString() }}</time>
              </template>
              <template v-else-if="column.key === 'action'">
                <!-- Removed the div wrapper to allow default left alignment -->
                <a-button type="link" @click="openCertificateModal(record.id)">管理证书</a-button>
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

    <!-- 创建汽车的对话框 -->
    <a-modal
      v-model:visible="showCreateModal"
      title="登记新汽车"
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
        <a-form-item label="车型" name="model" extra="请输入汽车的品牌和型号">
          <a-input-group compact>
            <a-input
              v-model:value="formState.model"
              placeholder="例如: 特斯拉 Model 3"
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

        <a-form-item label="车辆识别代号 (VIN)" name="vin" extra="请输入17位车辆识别代号">
          <a-input-group compact>
            <a-input
              v-model:value="formState.vin"
              placeholder="请输入VIN"
              style="width: calc(100% - 110px)"
            />
             <a-tooltip title="随机生成一个VIN">
              <a-button @click="generateRandomVINHandler">
                <template #icon><ReloadOutlined /></template>
                随机生成
              </a-button>
            </a-tooltip>
          </a-input-group>
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
          <span>汽车ID将由系统自动生成</span>
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
            :show-total="(total: number) => `共 ${total} 条`"
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

    <!-- 证书管理对话框 (新增) -->
    <a-modal
      v-model:visible="showCertificateModal"
      :title="`管理汽车 ${currentCarIdForCert} 的证书`"
      :confirmLoading="certificateModalLoading"
      :footer="null"
      @cancel="handleCertificateModalCancel"
      width="800px"
    >
      <div class="certificate-modal-content">
        <a-card title="已上传证书" :bordered="false" class="modal-card">
          <a-list
            :loading="certificateModalLoading"
            :data-source="certificateList"
            item-layout="horizontal"
            class="certificate-list"
          >
            <template #renderItem="{ item }">
              <a-list-item class="cert-list-item">
                <a-list-item-meta>
                  <template #avatar>
                    <!-- Ensure FileTextOutlined is imported -->
                    <a-avatar><FileTextOutlined /></a-avatar>
                  </template>
                  <template #title>
                    <a :href="`/api/files/${item.fileLocation.replace(/^data\//, '')}`" target="_blank" class="cert-title">
                      {{ item.certType === 'REGISTRATION' ? '登记证书' : '其他证书' }}
                    </a>
                    <span class="cert-id"> (ID: {{ item.certId }})</span>
                  </template>
                  <template #description>
                  上传时间: {{ new Date(item.uploadTime).toLocaleString() }}
                  <!-- Removed Hash Display -->
                </template>
              </a-list-item-meta>
              <template #actions>
                <a-tooltip title="验证服务器存储的文件"> <!-- Clarify tooltip -->
                  <a-button
                    type="text"
                    shape="circle"
                      :loading="verificationLoading[item.certId]"
                      @click="handleVerifyCertificate(item.certId)"
                      class="verify-button"
                    >
                      <!-- Ensure CheckCircleOutlined is imported -->
                      <template #icon><CheckCircleOutlined /></template>
                    </a-button>
                  </a-tooltip>
                </template>
              </a-list-item>
            </template>
            <template #empty>
              <a-empty description="暂无证书记录" />
            </template>
          </a-list>
        </a-card>

        <a-card title="上传新证书" :bordered="false" class="modal-card">
          <!-- Conditionally disable upload section -->
          <div v-if="certificateList.length > 0" class="upload-disabled-message">
            <a-alert message="每辆车只能上传一份证书。" type="info" show-icon />
          </div>
          <a-space v-else align="baseline">
            <a-select
              v-model:value="selectedCertType"
              style="width: 150px;"
            >
              <a-select-option value="REGISTRATION">登记证书</a-select-option>
              <a-select-option value="OTHER">其他证书</a-select-option>
            </a-select>
            <a-upload
              :file-list="certificateFile ? [certificateFile] : []"
              :before-upload="() => false"
              @change="handleFileChange"
              :max-count="1"
            >
              <a-button>
                <!-- Ensure UploadOutlined is imported -->
                <template #icon><UploadOutlined /></template>
                选择文件
              </a-button>
            </a-upload>
            <a-button
              type="primary"
              :disabled="!certificateFile"
              :loading="certificateModalLoading"
              @click="handleCertificateUpload"
            >
              上传
            </a-button>
          </a-space>
        </a-card>

        <!-- Verification Upload Card (Only show if a certificate exists) -->
        <a-card v-if="certificateList.length > 0" title="上传文件进行验证" :bordered="false" class="modal-card">
          <a-space align="baseline">
            <a-upload
              :file-list="verificationFile ? [verificationFile] : []"
              :before-upload="() => false"
              @change="handleVerificationFileChange"
              :max-count="1"
            >
              <a-button>
                <template #icon><UploadOutlined /></template>
                选择要验证的文件
              </a-button>
            </a-upload>
            <a-button
              type="primary"
              :disabled="!verificationFile"
              :loading="verificationUploadLoading"
              @click="handleVerificationUpload"
            >
              验证上传文件
            </a-button>
          </a-space>
        </a-card>
      </div>
    </a-modal>

  </div>
</template>

<script setup lang="ts">
import { message, UploadChangeParam, UploadFile, Modal } from 'ant-design-vue'; // Import Upload types and Modal
import { PlusOutlined, InfoCircleOutlined, ReloadOutlined, CopyOutlined, ApartmentOutlined, UploadOutlined, CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined } from '@ant-design/icons-vue'; // Add icons
import { carDealerApi } from '../api';
import type { FormInstance } from 'ant-design-vue';
import { ref, reactive, watch, onMounted } from 'vue';
import type { BlockData, Car, Certificate } from '../types'; // Import Car and Certificate types
import { copyToClipboard, generateRandomName, generateUUID, generateRandomCarModel, generateRandomVIN } from '../utils';

const formRef = ref<FormInstance>();
const showCreateModal = ref(false);
const modalLoading = ref(false);

const formState = reactive({
  model: '',
  vin: '',
  owner: '',
});

const rules = {
  model: [{ required: true, message: '请输入车型' }],
  vin: [
    { required: true, message: '请输入车辆识别代号 (VIN)' },
    { len: 17, message: 'VIN 必须是17位' }
  ],
  owner: [{ required: true, message: '请输入所有者' }],
};

const columns = [
  {
    title: '汽车ID',
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
    title: '车型',
    dataIndex: 'model',
    key: 'model',
    width: 120,
    ellipsis: { showTitle: true },
  },
  {
    title: 'VIN',
    dataIndex: 'vin',
    key: 'vin',
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
    title: '当前所有者',
    dataIndex: 'currentOwner',
    key: 'currentOwner',
    width: 120,
    ellipsis: false,
    customCell: () => ({
      style: {
        whiteSpace: 'nowrap',
        overflow: 'hidden',
      }
    }),
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
  {
    title: '操作',
    key: 'action',
    width: 120,
    fixed: 'right' as 'right', // Explicitly cast to fix TS error
  },
];

const carList = ref<Car[]>([]); // Use Car type
const loading = ref(false);
const bookmark = ref('');

// --- Certificate State ---
const showCertificateModal = ref(false);
const certificateModalLoading = ref(false);
const currentCarIdForCert = ref('');
const certificateList = ref<Certificate[]>([]);
const certificateFile = ref<UploadFile | null>(null);
const selectedCertType = ref<'REGISTRATION' | 'OTHER'>('REGISTRATION'); // Default type
const verificationLoading = ref<Record<string, boolean>>({}); // Loading state for each verify button (新增)
const verificationFile = ref<UploadFile | null>(null); // State for verification upload (新增)
const verificationUploadLoading = ref(false); // Loading state for verification upload (新增)

const loadCarList = async () => {
  try {
    loading.value = true;
    const result = await carDealerApi.getCarList({
      pageSize: 10,
      bookmark: bookmark.value,
      status: statusFilter.value,
    });
    if (!bookmark.value) {
      carList.value = result.records;
    } else {
      carList.value = [...carList.value, ...result.records];
    }
    bookmark.value = result.bookmark;
  } catch (error: any) {
    message.error(error.message || '加载汽车列表失败');
  } finally {
    loading.value = false;
  }
};

const loadMore = () => {
  loadCarList();
};

const handleModalOk = () => {
  formRef.value?.validate().then(async () => {
    modalLoading.value = true;
    try {
      const carData = {
        ...formState,
        id: generateUUID(),
      };
      await carDealerApi.createCar(carData);
      message.success('汽车信息登记成功');
      showCreateModal.value = false;
      formRef.value?.resetFields();
      carList.value = [];
      bookmark.value = '';
      loadCarList();
    } catch (error: any) {
      message.error(error.message || '汽车信息登记失败');
    } finally {
      modalLoading.value = false;
    }
  });
};

const handleModalCancel = () => {
  showCreateModal.value = false;
  formRef.value?.resetFields();
};

const generateRandomModelHandler = () => {
  formState.model = generateRandomCarModel();
};

const generateRandomVINHandler = () => {
  formState.vin = generateRandomVIN();
};

const generateRandomOwner = () => {
  formState.owner = generateRandomName();
};

const handleCopy = (text: string) => {
  copyToClipboard(text);
};

const statusFilter = ref('');

watch(statusFilter, () => {
  carList.value = [];
  bookmark.value = '';
  loadCarList();
});

const searchId = ref('');

const handleSearch = async (value: string) => {
  if (!value) {
    message.warning('请输入要查询的汽车ID');
    return;
  }

  try {
    const result = await carDealerApi.getCar(value);
    carList.value = [result];
    bookmark.value = '';
  } catch (error: any) {
    message.error(error.message || '查询汽车信息失败');
    carList.value = [];
  }
};

const handleSearchChange = (e: Event) => {
  const value = (e.target as HTMLInputElement).value;
  if (!value) {
    carList.value = [];
    bookmark.value = '';
    loadCarList();
  }
};

onMounted(() => {
  loadCarList();
});

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
    const res = await carDealerApi.getBlockList({
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

// --- Certificate Functions (新增) ---

// Fetch certificates for a specific car
const fetchCertificates = async (carId: string) => {
  try {
    certificateModalLoading.value = true;
    const result = await carDealerApi.listCertificates(carId);
    // Ensure certificateList is always an array, even if API returns null/undefined
    certificateList.value = Array.isArray(result) ? result : [];
  } catch (error: any) {
    message.error(error.message || '获取证书列表失败');
    certificateList.value = []; // Also ensure empty array on error
  } finally {
    certificateModalLoading.value = false;
  }
};

// Open the certificate management modal
const openCertificateModal = async (carId: string) => {
  currentCarIdForCert.value = carId;
  showCertificateModal.value = true;
  certificateFile.value = null; // Reset file input
  selectedCertType.value = 'REGISTRATION'; // Reset type
  await fetchCertificates(carId); // Fetch certificates when modal opens
};

// Handle modal cancellation
const handleCertificateModalCancel = () => {
  showCertificateModal.value = false;
  certificateList.value = []; // Clear list
  currentCarIdForCert.value = '';
};

// Handle file selection change
const handleFileChange = (info: UploadChangeParam) => {
  // Keep only the last selected file
  if (info.fileList.length > 0) {
    certificateFile.value = info.fileList[info.fileList.length - 1];
  } else {
    certificateFile.value = null;
  }
};

// Handle certificate upload
const handleCertificateUpload = async () => {
  if (!certificateFile.value?.originFileObj) {
    message.error('请先选择要上传的文件');
    return;
  }
  if (!currentCarIdForCert.value) {
    message.error('无法确定要为哪辆车上传证书');
    return;
  }

  try {
    certificateModalLoading.value = true;
    await carDealerApi.uploadCertificate(
      currentCarIdForCert.value,
      selectedCertType.value,
      certificateFile.value.originFileObj as File // Pass the actual File object
    );
    message.success('证书上传成功');
    certificateFile.value = null; // Clear file input after successful upload
    await fetchCertificates(currentCarIdForCert.value); // Refresh the list
  } catch (error: any) {
    message.error(error.message || '证书上传失败');
  } finally {
    certificateModalLoading.value = false;
  }
};

// Handle certificate verification (新增)
const handleVerifyCertificate = async (certId: string) => {
  verificationLoading.value[certId] = true;
  try {
    const result = await carDealerApi.verifyCertificate(certId);
    if (result.match) {
      Modal.success({
        title: '验证成功',
        content: `文件哈希与链上记录一致。\n链上哈希: ${result.storedHash}\n当前哈希: ${result.currentHash}`,
        width: 600, // Wider modal for hash display
      });
    } else {
      Modal.error({
        title: '验证失败',
        content: `文件哈希与链上记录不一致！文件可能已被修改。\n链上哈希: ${result.storedHash}\n当前哈希: ${result.currentHash}`,
        width: 600,
      });
    }
  } catch (error: any) {
    message.error(error.message || '验证证书时发生错误');
  } finally {
    verificationLoading.value[certId] = false;
  }
};

// --- Verification Upload Functions (新增) ---

// Handle verification file selection change
const handleVerificationFileChange = (info: UploadChangeParam) => {
  if (info.fileList.length > 0) {
    verificationFile.value = info.fileList[info.fileList.length - 1];
  } else {
    verificationFile.value = null;
  }
};

// Handle verification upload and comparison
const handleVerificationUpload = async () => {
  if (!verificationFile.value?.originFileObj) {
    message.error('请先选择要验证的文件');
    return;
  }
  if (!currentCarIdForCert.value) {
    message.error('无法确定要为哪辆车验证证书');
    return;
  }

  verificationUploadLoading.value = true;
  try {
    const result = await carDealerApi.verifyUploadedCertificate(
      currentCarIdForCert.value,
      verificationFile.value.originFileObj as File
    );

    // Display result in a modal similar to handleVerifyCertificate
    if (result.match) {
      Modal.success({
        title: '验证成功',
        content: `上传的文件与原始证书哈希一致。\n原始哈希: ${result.storedHash}\n上传文件哈希: ${result.currentHash}`,
        width: 600,
      });
    } else {
      Modal.error({
        title: '验证失败',
        content: `上传的文件与原始证书哈希不一致！\n原始哈希: ${result.storedHash}\n上传文件哈希: ${result.currentHash}`,
        width: 600,
      });
    }
    verificationFile.value = null; // Clear file input after verification attempt
  } catch (error: any) {
    // Handle specific errors like "no original certificate"
    if (error.message && error.message.includes("没有已上传的原始证书记录")) {
       message.error('此车辆没有原始证书记录，无法进行比对。');
    } else {
       message.error(error.message || '验证上传文件时发生错误');
    }
  } finally {
    verificationUploadLoading.value = false;
  }
};

</script>

<style scoped>
.car-dealer {
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
  max-width: 120px;
  display: inline-block;
  vertical-align: middle;
}

.custom-table .copy-icon {
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
  padding-right: 8px;
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
  min-width: 70px;
}

.field-value {
  flex: 1;
  word-break: break-all;
}

.field-value.hash {
  font-family: 'Courier New', Courier, monospace;
  max-width: calc(100% - 100px);
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

/* --- Certificate Modal Styles --- */
.certificate-modal-content {
  padding: 0 8px; /* Add some padding */
}

.modal-card {
  margin-bottom: 24px; /* Space between cards */
}

.certificate-list .ant-list-item {
  padding-top: 16px;
  padding-bottom: 16px;
}

.cert-list-item .ant-list-item-meta-avatar {
  align-self: flex-start; /* Align avatar to top */
  margin-top: 4px;
}

.cert-title {
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
}

.cert-id {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
  margin-left: 4px;
}

.cert-hash-label {
  font-weight: 500;
  color: rgba(0, 0, 0, 0.65);
  margin-right: 8px;
}

.cert-hash {
  font-family: 'Courier New', Courier, monospace;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.65);
  max-width: 400px; /* Limit width */
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  vertical-align: bottom; /* Align with label */
}

.verify-button.ant-btn-circle {
  border: none;
  box-shadow: none;
}
.verify-button.ant-btn-loading {
   background-color: transparent !important; /* Prevent background on loading */
}

</style>
