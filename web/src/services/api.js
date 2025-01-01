import axios from 'axios';
import config from '../config';

// Set base URL for all requests
axios.defaults.baseURL = config.apiBaseUrl;

// System APIs
export const getSystemInfo = async () => {
  const response = await axios.get('/api/system/info');
  return response.data;
};

export const getSystemStatus = async () => {
  const response = await axios.get('/api/system/status');
  return response.data;
};

export const startService = async (name) => {
  const response = await axios.post(`/api/system/services/${name}/start`);
  return response.data;
};

export const stopService = async (name) => {
  const response = await axios.post(`/api/system/services/${name}/stop`);
  return response.data;
};

export const restartService = async (name) => {
  const response = await axios.post(`/api/system/services/${name}/restart`);
  return response.data;
};

// Node APIs
export const getNodes = (page = 1, pageSize = 10) => {
  return axios.get('/api/nodes', {
    params: {
      page,
      page_size: pageSize
    }
  }).then(response => response.data);
};

export const testNodes = async () => {
  const response = await axios.post('/api/nodes/test');
  return response.data;
};

export const getNodeStatus = async (id) => {
  const response = await axios.get(`/api/nodes/${id}/status`);
  return response.data;
};

export const createNode = (node) => {
  return axios.post('/api/nodes', node).then(response => response.data);
};

export const updateNode = (id, node) => {
  return axios.put(`/api/nodes/${id}`, node).then(response => response.data);
};

export const deleteNode = (id) => {
  return axios.delete(`/api/nodes/${id}`).then(response => response.data);
};

export const importNodes = async (url, autoUpdate = false, updateInterval = 0) => {
  const response = await axios.post('/api/nodes/import', {
    url,
    auto_update: autoUpdate,
    update_interval: updateInterval
  });
  return response.data;
};

export const testNode = (id) => {
  return axios.post(`/api/nodes/${id}/test`).then(response => response.data);
};

// Rule APIs
export const getRules = async () => {
  const response = await axios.get('/api/rules');
  return response.data;
};

export const createRule = async (rule) => {
  const response = await axios.post('/api/rules', rule);
  return response.data;
};

export const updateRule = async (id, rule) => {
  const response = await axios.put(`/api/rules/${id}`, rule);
  return response.data;
};

export const deleteRule = async (id) => {
  const response = await axios.delete(`/api/rules/${id}`);
  return response.data;
};

// Subscription APIs
export const getSubscriptions = async () => {
  const response = await axios.get('/api/subscriptions');
  return response.data;
};

export const createSubscription = async (subscription) => {
  const response = await axios.post('/api/subscriptions', subscription);
  return response.data;
};

export const updateSubscription = async (id, subscription) => {
  const response = await axios.put(`/api/subscriptions/${id}`, subscription);
  return response.data;
};

export const deleteSubscription = async (id) => {
  const response = await axios.delete(`/api/subscriptions/${id}`, {
    params: {
      delete_nodes: true
    }
  });
  return response.data;
};

export const updateSubscriptionNodes = async (id) => {
  const response = await axios.post(`/api/subscriptions/${id}/update`);
  return response.data;
};

export const refreshSubscription = async (id) => {
  const response = await axios.post(`/api/subscriptions/${id}/refresh`);
  return response.data;
};

// Settings APIs
export const getSettings = async () => {
  const response = await axios.get('/api/settings');
  return response.data;
};

export const updateSettings = async (settings) => {
  const response = await axios.put('/api/settings', settings);
  return response.data;
};

export const updatePassword = async (data) => {
  const response = await axios.put('/api/settings/password', data);
  return response.data;
};

// Traffic APIs
export const getTrafficStats = async () => {
  const response = await axios.get('/api/traffic/stats');
  return response.data;
};

export const getRealtimeTraffic = async () => {
  const response = await axios.get('/api/traffic/realtime');
  return response.data;
};

// Node Groups
export const getNodeGroups = () => {
  return axios.get('/api/node-groups')
    .then(response => response.data);
};

export const createNodeGroup = (group) => {
  console.log('Creating node group with data:', {
    ...group,
    includePatterns: JSON.stringify(group.includePatterns),
    excludePatterns: JSON.stringify(group.excludePatterns)
  });
  return axios.post('/api/node-groups', group)
    .then(response => response.data);
};

export const updateNodeGroup = (id, group) => {
  console.log('Updating node group with data:', group);
  return axios.put(`/api/node-groups/${id}`, group)
    .then(response => response.data);
};

export const deleteNodeGroup = async (id) => {
  const response = await axios.delete(`/api/node-groups/${id}`);
  return response.data;
};

// Rule set API calls
export const getRuleSets = async () => {
  const response = await axios.get('/api/config/rulesets');
  return response.data;
};

export const createRuleSet = async (data) => {
  const response = await axios.post('/api/config/rulesets', data);
  return response.data;
};

export const updateRuleSet = async (id, data) => {
  const response = await axios.put(`/api/config/rulesets/${id}`, data);
  return response.data;
};

export const deleteRuleSet = async (id) => {
  const response = await axios.delete(`/api/config/rulesets/${id}`);
  return response.data;
};

export const updateRuleSetRules = async (id) => {
  const response = await axios.post(`/api/config/rulesets/${id}/update`);
  return response.data;
};

// DNS 规则管理
export const getDNSRules = () => axios.get('/api/dns/rules');
export const createDNSRule = (data) => axios.post('/api/dns/rules', data);
export const updateDNSRule = (id, data) => axios.put(`/api/dns/rules/${id}`, data);
export const deleteDNSRule = (id) => axios.delete(`/api/dns/rules/${id}`);
export const updateDNSSettings = (data) => axios.put('/api/dns/settings', data);

// Hosts 文件管理
export const getHosts = () => axios.get('/api/config/hosts');
export const createHost = (data) => axios.post('/api/config/hosts', data);
export const updateHost = (id, data) => axios.put(`/api/config/hosts/${id}`, data);
export const deleteHost = (id) => axios.delete(`/api/config/hosts/${id}`);

// Add axios interceptors for error handling
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle 401 Unauthorized
    if (error.response?.status === 401) {
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// 更新规则集
export const updateRuleSets = async () => {
  const response = await axios.post('/api/config/rulesets/update');
  return response.data;
};

// 生成配置文件
export const generateConfigs = async () => {
  const response = await axios.post('/api/config/generate');
  return response.data;
};
 