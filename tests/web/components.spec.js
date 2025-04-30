import { mount, shallowMount } from '@vue/test-utils';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { message } from 'ant-design-vue';
import CarDealer from '../../application/web/src/views/CarDealer.vue';
import TradingPlatform from '../../application/web/src/views/TradingPlatform.vue';
import Bank from '../../application/web/src/views/Bank.vue';
import { carDealerApi, tradingPlatformApi, bankApi } from '../../application/web/src/api';

// 模拟 API
vi.mock('../../application/web/src/api', () => {
  return {
    carDealerApi: {
      createCar: vi.fn(),
      getCar: vi.fn(),
      getCarList: vi.fn(),
      uploadCertificate: vi.fn(),
      listCertificates: vi.fn(),
      verifyCertificate: vi.fn(),
      getBlockList: vi.fn(),
    },
    tradingPlatformApi: {
      createTransaction: vi.fn(),
      getCar: vi.fn(),
      getTransaction: vi.fn(),
      getTransactionList: vi.fn(),
      getBlockList: vi.fn(),
    },
    bankApi: {
      completeTransaction: vi.fn(),
      getTransaction: vi.fn(),
      getTransactionList: vi.fn(),
      getBlockList: vi.fn(),
    },
  };
});

// 模拟 ant-design-vue 组件
vi.mock('ant-design-vue', () => {
  return {
    message: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
    },
  };
});

// 模拟 Vue Router
vi.mock('vue-router', () => ({
  createRouter: vi.fn(() => ({
    push: vi.fn(),
  })),
  createWebHistory: vi.fn(),
}));

// 模拟工具函数
vi.mock('../../application/web/src/utils', () => {
  return {
    copyToClipboard: vi.fn(),
    generateRandomName: vi.fn(() => 'RandomName'),
    generateUUID: vi.fn(() => 'UUID123'),
    generateRandomCarModel: vi.fn(() => '特斯拉 Model 3'),
    generateRandomVIN: vi.fn(() => 'VIN123456789ABCDE'),
  };
});

describe('CarDealer.vue', () => {
  let wrapper;

  beforeEach(() => {
    // 重置模拟函数
    vi.clearAllMocks();

    // 模拟 API 响应
    carDealerApi.getCarList.mockResolvedValue({
      records: [
        {
          id: 'CAR001',
          model: '特斯拉 Model 3',
          vin: 'VIN123456789ABCDE',
          currentOwner: 'Alice',
          status: 'AVAILABLE',
          createTime: '2023-01-01T00:00:00Z',
          updateTime: '2023-01-01T00:00:00Z',
        },
      ],
      bookmark: '',
      recordsCount: 1,
      fetchedRecordsCount: 1,
    });

    carDealerApi.getBlockList.mockResolvedValue({
      blocks: [
        {
          block_num: 1,
          block_hash: 'hash1',
          data_hash: 'data_hash1',
          prev_hash: 'prev_hash1',
          tx_count: 1,
          save_time: '2023-01-01T00:00:00Z',
        },
      ],
      total: 1,
    });

    // 浅挂载组件
    wrapper = shallowMount(CarDealer);
  });

  it('加载汽车列表', async () => {
    // 等待异步操作完成
    await vi.runAllTimersAsync();

    // 验证 API 调用
    expect(carDealerApi.getCarList).toHaveBeenCalledWith({
      pageSize: 10,
      bookmark: '',
      status: '',
    });

    // 验证数据是否正确加载
    expect(wrapper.vm.carList.length).toBe(1);
    expect(wrapper.vm.carList[0].id).toBe('CAR001');
  });

  it('创建汽车', async () => {
    // 模拟 API 响应
    carDealerApi.createCar.mockResolvedValue({});

    // 打开创建汽车对话框
    wrapper.vm.showCreateModal = true;
    await wrapper.vm.$nextTick();

    // 设置表单数据
    wrapper.vm.formState.model = '特斯拉 Model 3';
    wrapper.vm.formState.vin = 'VIN123456789ABCDE';
    wrapper.vm.formState.owner = 'Alice';

    // 模拟表单验证通过
    wrapper.vm.formRef = {
      validate: vi.fn().mockResolvedValue(true),
      resetFields: vi.fn(),
    };

    // 提交表单
    await wrapper.vm.handleModalOk();

    // 验证 API 调用
    expect(carDealerApi.createCar).toHaveBeenCalledWith({
      id: expect.any(String),
      model: '特斯拉 Model 3',
      vin: 'VIN123456789ABCDE',
      owner: 'Alice',
    });

    // 验证成功消息
    expect(message.success).toHaveBeenCalledWith('汽车信息登记成功');

    // 验证对话框关闭
    expect(wrapper.vm.showCreateModal).toBe(false);
  });

  it('上传证书', async () => {
    // 模拟 API 响应
    carDealerApi.uploadCertificate.mockResolvedValue({});
    carDealerApi.listCertificates.mockResolvedValue([]);

    // 打开证书管理对话框
    await wrapper.vm.openCertificateModal('CAR001');

    // 验证 API 调用
    expect(carDealerApi.listCertificates).toHaveBeenCalledWith('CAR001');

    // 设置证书数据
    wrapper.vm.selectedCertType = 'REGISTRATION';
    wrapper.vm.certificateFile = {
      originFileObj: new File([''], 'cert.pdf', { type: 'application/pdf' }),
    };

    // 上传证书
    await wrapper.vm.handleCertificateUpload();

    // 验证 API 调用
    expect(carDealerApi.uploadCertificate).toHaveBeenCalledWith(
      'CAR001',
      'REGISTRATION',
      expect.any(File)
    );

    // 验证成功消息
    expect(message.success).toHaveBeenCalledWith('证书上传成功');
  });
});

describe('TradingPlatform.vue', () => {
  let wrapper;

  beforeEach(() => {
    // 重置模拟函数
    vi.clearAllMocks();

    // 模拟 API 响应
    tradingPlatformApi.getTransactionList.mockResolvedValue({
      records: [
        {
          id: 'TX001',
          carId: 'CAR001',
          seller: 'Alice',
          buyer: 'Bob',
          price: 100000,
          status: 'PENDING',
          createTime: '2023-01-01T00:00:00Z',
          updateTime: '2023-01-01T00:00:00Z',
        },
      ],
      bookmark: '',
      recordsCount: 1,
      fetchedRecordsCount: 1,
    });

    tradingPlatformApi.getCar.mockResolvedValue({
      id: 'CAR001',
      model: '特斯拉 Model 3',
      vin: 'VIN123456789ABCDE',
      currentOwner: 'Alice',
      status: 'AVAILABLE',
      createTime: '2023-01-01T00:00:00Z',
      updateTime: '2023-01-01T00:00:00Z',
    });

    // 浅挂载组件
    wrapper = shallowMount(TradingPlatform);
  });

  it('加载交易列表', async () => {
    // 等待异步操作完成
    await vi.runAllTimersAsync();

    // 验证 API 调用
    expect(tradingPlatformApi.getTransactionList).toHaveBeenCalledWith({
      pageSize: 10,
      bookmark: '',
      status: '',
    });

    // 验证数据是否正确加载
    expect(wrapper.vm.transactionList.length).toBe(1);
    expect(wrapper.vm.transactionList[0].id).toBe('TX001');
  });

  it('创建交易', async () => {
    // 模拟 API 响应
    tradingPlatformApi.createTransaction.mockResolvedValue({});

    // 打开创建交易对话框
    wrapper.vm.showCreateModal = true;
    await wrapper.vm.$nextTick();

    // 设置表单数据
    wrapper.vm.formState.carId = 'CAR001';
    wrapper.vm.formState.seller = 'Alice';
    wrapper.vm.formState.buyer = 'Bob';
    wrapper.vm.formState.price = 100000;

    // 模拟表单验证通过
    wrapper.vm.formRef = {
      validate: vi.fn().mockResolvedValue(true),
      resetFields: vi.fn(),
    };

    // 提交表单
    await wrapper.vm.handleModalOk();

    // 验证 API 调用
    expect(tradingPlatformApi.createTransaction).toHaveBeenCalledWith({
      id: expect.any(String),
      carId: 'CAR001',
      seller: 'Alice',
      buyer: 'Bob',
      price: 100000,
    });

    // 验证成功消息
    expect(message.success).toHaveBeenCalledWith('交易创建成功');

    // 验证对话框关闭
    expect(wrapper.vm.showCreateModal).toBe(false);
  });
});

describe('Bank.vue', () => {
  let wrapper;

  beforeEach(() => {
    // 重置模拟函数
    vi.clearAllMocks();

    // 模拟 API 响应
    bankApi.getTransactionList.mockResolvedValue({
      records: [
        {
          id: 'TX001',
          carId: 'CAR001',
          seller: 'Alice',
          buyer: 'Bob',
          price: 100000,
          status: 'PENDING',
          createTime: '2023-01-01T00:00:00Z',
          updateTime: '2023-01-01T00:00:00Z',
        },
      ],
      bookmark: '',
      recordsCount: 1,
      fetchedRecordsCount: 1,
    });

    // 浅挂载组件
    wrapper = shallowMount(Bank);
  });

  it('加载交易列表', async () => {
    // 等待异步操作完成
    await vi.runAllTimersAsync();

    // 验证 API 调用
    expect(bankApi.getTransactionList).toHaveBeenCalledWith({
      pageSize: 10,
      bookmark: '',
      status: '',
    });

    // 验证数据是否正确加载
    expect(wrapper.vm.transactionList.length).toBe(1);
    expect(wrapper.vm.transactionList[0].id).toBe('TX001');
  });

  it('完成交易', async () => {
    // 模拟 API 响应
    bankApi.completeTransaction.mockResolvedValue({});

    // 模拟确认对话框
    const originalModal = global.Modal;
    global.Modal = {
      confirm: ({ onOk }) => {
        onOk();
      },
    };

    // 完成交易
    await wrapper.vm.handleCompleteTransaction('TX001');

    // 验证 API 调用
    expect(bankApi.completeTransaction).toHaveBeenCalledWith('TX001');

    // 验证成功消息
    expect(message.success).toHaveBeenCalledWith('交易完成成功');

    // 恢复原始 Modal
    global.Modal = originalModal;
  });
});
