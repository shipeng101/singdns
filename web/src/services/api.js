import axios from 'axios';
import { getToken } from './auth';

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Add request interceptor
api.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Redirect to login page
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// System
export const getSystemStatus = () => api.get('/system/status');
export const getSystemVersion = () => api.get('/system/version');
export const restartService = () => api.post('/system/restart');

// Settings
export const getSettings = () => api.get('/settings');
export const updateSettings = (settings) => api.put('/settings', settings);

// Nodes
export const getNodes = () => api.get('/nodes');
export const createNode = (node) => api.post('/nodes', node);
export const updateNode = (id, node) => api.put(`/nodes/${id}`, node);
export const deleteNode = (id) => api.delete(`/nodes/${id}`);
export const testNode = (id) => api.post(`/nodes/${id}/test`);

// Node Groups
export const getNodeGroups = () => api.get('/node-groups');
export const createNodeGroup = (group) => api.post('/node-groups', group);
export const updateNodeGroup = (id, group) => api.put(`/node-groups/${id}`, group);
export const deleteNodeGroup = (id) => api.delete(`/node-groups/${id}`);

// Groups
export const getGroups = () => api.get('/groups');
export const createGroup = (group) => api.post('/groups', group);
export const updateGroup = (id, group) => api.put(`/groups/${id}`, group);
export const deleteGroup = (id) => api.delete(`/groups/${id}`);

// Rules
export const getRules = () => api.get('/rules');
export const createRule = (rule) => api.post('/rules', rule);
export const updateRule = (id, rule) => api.put(`/rules/${id}`, rule);
export const deleteRule = (id) => api.delete(`/rules/${id}`);
export const toggleRule = (id) => api.post(`/rules/${id}/toggle`);

// Subscriptions
export const getSubscriptions = () => api.get('/subscriptions');
export const createSubscription = (subscription) => api.post('/subscriptions', subscription);
export const updateSubscription = (id, subscription) => api.put(`/subscriptions/${id}`, subscription);
export const deleteSubscription = (id) => api.delete(`/subscriptions/${id}`);
export const updateSubscriptionNodes = (id) => api.post(`/subscriptions/${id}/update`);

export default api; 