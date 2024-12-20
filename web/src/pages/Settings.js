import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Switch,
  FormControlLabel,
  Snackbar,
  Alert,
  Grid,
  LinearProgress,
  Divider
} from '@mui/material';
import * as api from '../services/api';

function Settings() {
  const [settings, setSettings] = useState({
    httpPort: '',
    socksPort: '',
    apiPort: '',
    allowLan: false,
    logLevel: 'info'
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  const fetchSettings = async () => {
    setLoading(true);
    try {
      const response = await api.getSettings();
      setSettings(response.data);
    } catch (err) {
      setError(err.response?.data?.message || '获取设置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  const handleChange = (name) => (event) => {
    const value = event.target.type === 'checkbox' ? event.target.checked : event.target.value;
    setSettings(prev => ({ ...prev, [name]: value }));
  };

  const handleSave = async () => {
    setLoading(true);
    setError(null);
    try {
      await api.updateSettings(settings);
      setSuccess('设置已保存');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err.response?.data?.message || '保存设置失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box>
      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h5">系统设置</Typography>
        <Button
          variant="contained"
          onClick={handleSave}
          disabled={loading}
        >
          保存设置
        </Button>
      </Box>

      {loading && <LinearProgress sx={{ mb: 2 }} />}

      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Card sx={{
            '&:hover': {
              boxShadow: (theme) => theme.shadows[4],
              transform: 'translateY(-2px)',
              transition: 'all 0.3s'
            }
          }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                网络设置
              </Typography>
              <Divider sx={{ my: 2 }} />
              <Grid container spacing={2}>
                <Grid item xs={12} sm={4}>
                  <TextField
                    label="HTTP 代理端口"
                    type="number"
                    fullWidth
                    value={settings.httpPort}
                    onChange={handleChange('httpPort')}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <TextField
                    label="SOCKS 代理端口"
                    type="number"
                    fullWidth
                    value={settings.socksPort}
                    onChange={handleChange('socksPort')}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <TextField
                    label="API 端口"
                    type="number"
                    fullWidth
                    value={settings.apiPort}
                    onChange={handleChange('apiPort')}
                  />
                </Grid>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={settings.allowLan}
                        onChange={handleChange('allowLan')}
                      />
                    }
                    label="允许局域网访问"
                  />
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Card sx={{
            '&:hover': {
              boxShadow: (theme) => theme.shadows[4],
              transform: 'translateY(-2px)',
              transition: 'all 0.3s'
            }
          }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                日志设置
              </Typography>
              <Divider sx={{ my: 2 }} />
              <TextField
                select
                label="日志级别"
                fullWidth
                value={settings.logLevel}
                onChange={handleChange('logLevel')}
                SelectProps={{
                  native: true
                }}
              >
                <option value="debug">调试</option>
                <option value="info">信息</option>
                <option value="warning">警告</option>
                <option value="error">错误</option>
              </TextField>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Snackbar
        open={!!error}
        autoHideDuration={6000}
        onClose={() => setError(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setError(null)} severity="error">
          {error}
        </Alert>
      </Snackbar>

      <Snackbar
        open={!!success}
        autoHideDuration={3000}
        onClose={() => setSuccess(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={() => setSuccess(null)} severity="success">
          {success}
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default Settings; 