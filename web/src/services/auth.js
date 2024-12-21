import axios from 'axios';

const TOKEN_KEY = 'singdns_token';

export const login = async (username, password) => {
  try {
    const response = await axios.post('/api/auth/login', { username, password });
    const { token } = response.data;
    localStorage.setItem(TOKEN_KEY, token);
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    return response.data;
  } catch (error) {
    throw new Error(error.response?.data?.error || '登录失败');
  }
};

export const register = async (username, password) => {
  try {
    const response = await axios.post('/api/auth/register', { username, password });
    return response.data;
  } catch (error) {
    throw new Error(error.response?.data?.error || '注册失败');
  }
};

export const logout = () => {
  localStorage.removeItem(TOKEN_KEY);
  delete axios.defaults.headers.common['Authorization'];
};

export const updatePassword = async (oldPassword, newPassword) => {
  try {
    const response = await axios.put('/api/user/password', { 
      old_password: oldPassword,
      new_password: newPassword 
    });
    return response.data;
  } catch (error) {
    throw new Error(error.response?.data?.error || '修改密码失败');
  }
};

export const getToken = () => {
  return localStorage.getItem(TOKEN_KEY);
};

export const isAuthenticated = () => {
  const token = getToken();
  return !!token;
};

// Add token to axios headers if it exists
const token = getToken();
if (token) {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
} 