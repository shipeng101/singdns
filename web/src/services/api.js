import axios from 'axios';

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
  const response = await axios.post(`/api/system/service/${name}/start`);
  return response.data;
};

export const stopService = async (name) => {
  const response = await axios.post(`/api/system/service/${name}/stop`);
  return response.data;
};

export const restartService = async (name) => {
  const response = await axios.post(`/api/system/service/${name}/restart`);
  return response.data;
};

// Node APIs
export const getNodes = async () => {
  const response = await axios.get('/api/nodes');
  return response.data;
};

export const getNodeStatus = async (id) => {
  const response = await axios.get(`/api/nodes/${id}/status`);
  return response.data;
};

export const createNode = async (node) => {
  const response = await axios.post('/api/nodes', node);
  return response.data;
};

export const updateNode = async (id, node) => {
  const response = await axios.put(`/api/nodes/${id}`, node);
  return response.data;
};

export const deleteNode = async (id) => {
  const response = await axios.delete(`/api/nodes/${id}`);
  return response.data;
};

export const importNodes = async (url, autoUpdate = false, updateInterval = 0) => {
  const response = await axios.post('/api/nodes/import', {
    url,
    auto_update: autoUpdate,
    update_interval: updateInterval
  });
  return response.data;
};

export const testNode = async (id) => {
  const response = await axios.post(`/api/nodes/${id}/test`);
  return response.data;
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
  const response = await axios.delete(`/api/subscriptions/${id}`);
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
export const getNodeGroups = async () => {
  const response = await axios.get('/api/node-groups');
  return response.data;
};

export const createNodeGroup = async (group) => {
  const response = await axios.post('/api/node-groups', group);
  return response.data;
};

export const updateNodeGroup = async (id, group) => {
  const response = await axios.put(`/api/node-groups/${id}`, group);
  return response.data;
};

export const deleteNodeGroup = async (id) => {
  const response = await axios.delete(`/api/node-groups/${id}`);
  return response.data;
};

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
 