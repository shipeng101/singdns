import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Grid,
  Button,
  Typography,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Stack,
  Alert,
  Snackbar,
  LinearProgress,
  useTheme,
  FormControlLabel,
  Switch,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Chip,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Update as UpdateIcon,
} from '@mui/icons-material';
import {
  getRules,
  createRule,
  updateRule,
  deleteRule,
  getNodeGroups,
  getHosts,
  createHost,
  updateHost,
  deleteHost,
  getDNSRules,
  createDNSRule,
  updateDNSRule,
  deleteDNSRule,
  updateDNSSettings,
  getRuleSets,
  updateRuleSets,
  createRuleSet,
  updateRuleSet,
  generateConfigs,
  deleteRuleSet,
} from '../services/api';
import { getCommonStyles } from '../styles/commonStyles';
import PageHeader from '../components/PageHeader';

// 统一的卡片样式
const cardStyle = {
  borderRadius: '12px',
  '& .MuiCardContent-root:last-child': {
    pb: 1.5
  }
};

const innerCardStyle = {
  borderRadius: '8px',
  '& .MuiCardContent-root:last-child': {
    pb: 1.5
  }
};

// SingBox 路由规则部分
const SingBoxRules = ({ rules, loading, onAdd, onEdit, onDelete, nodeGroups }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);

  const [ruleSets, setRuleSets] = useState([]);
  const [ruleSetError, setRuleSetError] = useState(null);
  const [ruleSetLoading, setRuleSetLoading] = useState({});
  const [adBlockEnabled, setAdBlockEnabled] = useState(true);

  // 加载规则集信息
  const loadRuleSets = async () => {
    try {
      const data = await getRuleSets();
      setRuleSets(data);
      // 从规则集中找到广告规则集的启用状态
      const adRuleSet = data.find(rs => rs.tag === 'geosite-category-ads');
      if (adRuleSet) {
        setAdBlockEnabled(adRuleSet.enabled);
      }
    } catch (error) {
      setRuleSetError('加载规则集信息失败');
    }
  };

  // 更新单个规则集
  const handleUpdate = async (tag) => {
    setRuleSetLoading(prev => ({ ...prev, [tag]: true }));
    try {
      await updateRuleSets();
      await loadRuleSets();
    } catch (error) {
      setRuleSetError('更新规则集失败');
    } finally {
      setRuleSetLoading(prev => ({ ...prev, [tag]: false }));
    }
  };

  // 删除规则集
  const handleDelete = async (id) => {
    setRuleSetLoading(prev => ({ ...prev, [id]: true }));
    try {
      await deleteRuleSet(id);
      await generateConfigs();
      await loadRuleSets();
    } catch (error) {
      setRuleSetError('删除规则集失败');
    } finally {
      setRuleSetLoading(prev => ({ ...prev, [id]: false }));
    }
  };

  // 处理广告规则集开关
  const handleAdBlockToggle = async (checked) => {
    setRuleSetLoading(prev => ({ ...prev, ads: true }));
    try {
      // 更新广告规则集状态
      const adRuleSet = ruleSets.find(rs => rs.tag === 'geosite-category-ads');
      if (adRuleSet) {
        await updateRuleSet(adRuleSet.id, {
          ...adRuleSet,
          enabled: checked
        });
        // 重新生成配置文件
        await generateConfigs();
        await loadRuleSets();
        setAdBlockEnabled(checked);
      }
    } catch (error) {
      setRuleSetError('更新广告规则集状态失败');
      console.error('Error updating ad-block rule set:', error);
    } finally {
      setRuleSetLoading(prev => ({ ...prev, ads: false }));
    }
  };

  useEffect(() => {
    loadRuleSets();
  }, []);

  // 分离内置规则集和用户规则集
  const geoRuleSets = ruleSets.filter(ruleSet => 
    ['geoip-cn', 'geosite-cn'].includes(ruleSet.tag)
  );
  const adRuleSets = ruleSets.filter(ruleSet => 
    ['geosite-category-ads'].includes(ruleSet.tag)
  );
  const userRuleSets = ruleSets.filter(ruleSet => 
    !['geoip-cn', 'geosite-cn', 'geosite-category-ads'].includes(ruleSet.tag)
  );

  return (
    <Box sx={{ mb: 3 }}>
      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
          SingBox 路由规则
        </Typography>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={onAdd}
          sx={styles.actionButton}
        >
          添加规则
        </Button>
      </Stack>

      <Grid container spacing={2}>
        {/* GeoIP 和 GeoSite 规则集管理卡片 */}
        <Grid item xs={12} sm={6} md={3}>
          <Card variant="outlined" sx={{ ...innerCardStyle, height: '100%' }}>
            <CardContent sx={{ p: 1.5 }}>
              <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                <Box sx={{ width: 'calc(100% - 64px)' }}>
                  <Typography variant="subtitle2" noWrap>GeoIP 和 GeoSite</Typography>
                  <Typography variant="caption" color="text.secondary">
                    中国 IP 和域名列表
                  </Typography>
                </Box>
                <Box>
                  <IconButton
                    size="small"
                    onClick={() => handleUpdate('geo')}
                    disabled={ruleSetLoading['geo']}
                    sx={{ p: 0.5 }}
                  >
                    <UpdateIcon fontSize="small" />
                  </IconButton>
                </Box>
              </Stack>

              <Stack direction="row" spacing={0.5} sx={{ mt: 1 }}>
                {geoRuleSets.map((ruleSet) => (
                  <Chip 
                    key={ruleSet.tag}
                    label={ruleSet.description}
                    size="small"
                    color="default"
                    sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                  />
                ))}
              </Stack>

              {geoRuleSets.length > 0 && (
                <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: 'block' }}>
                  最后更新：{new Date(geoRuleSets[0].updated_at).toLocaleString()}
                </Typography>
              )}

              {ruleSetLoading['geo'] && <LinearProgress sx={{ mt: 1 }} />}
            </CardContent>
          </Card>
        </Grid>

        {/* 广告规则集管理卡片 */}
        <Grid item xs={12} sm={6} md={3}>
          <Card variant="outlined" sx={{ ...innerCardStyle, height: '100%' }}>
            <CardContent sx={{ p: 1.5 }}>
              <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                <Box sx={{ width: 'calc(100% - 96px)' }}>
                  <Typography variant="subtitle2" noWrap>广告规则集</Typography>
                  <Typography variant="caption" color="text.secondary">
                    广告域名列表
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <FormControlLabel
                    control={
                      <Switch
                        size="small"
                        checked={adBlockEnabled}
                        onChange={(e) => handleAdBlockToggle(e.target.checked)}
                        disabled={ruleSetLoading['ads']}
                      />
                    }
                    label=""
                  />
                  <IconButton
                    size="small"
                    onClick={() => handleUpdate('ads')}
                    disabled={ruleSetLoading['ads']}
                    sx={{ p: 0.5 }}
                  >
                    <UpdateIcon fontSize="small" />
                  </IconButton>
                </Box>
              </Stack>

              <Stack direction="row" spacing={0.5} sx={{ mt: 1 }}>
                {adRuleSets.map((ruleSet) => (
                  <Chip 
                    key={ruleSet.tag}
                    label={ruleSet.description}
                    size="small"
                    color={adBlockEnabled ? "success" : "default"}
                    sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                  />
                ))}
              </Stack>

              {adRuleSets.length > 0 && (
                <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: 'block' }}>
                  最后更新：{new Date(adRuleSets[0].updated_at).toLocaleString()}
                </Typography>
              )}

              {ruleSetLoading['ads'] && <LinearProgress sx={{ mt: 1 }} />}
            </CardContent>
          </Card>
        </Grid>

        {/* 用户添加的规则集卡片 */}
        {userRuleSets.map((ruleSet) => (
          <Grid item xs={12} sm={6} md={3} key={ruleSet.tag}>
            <Card variant="outlined" sx={{ ...innerCardStyle, height: '100%' }}>
              <CardContent sx={{ p: 1.5 }}>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                  <Box sx={{ width: 'calc(100% - 96px)' }}>
                    <Typography variant="subtitle2" noWrap>{ruleSet.description}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      {ruleSet.tag}
                    </Typography>
                  </Box>
                  <Box>
                    <IconButton
                      size="small"
                      onClick={() => handleUpdate(ruleSet.tag)}
                      disabled={ruleSetLoading[ruleSet.tag]}
                      sx={{ p: 0.5 }}
                    >
                      <UpdateIcon fontSize="small" />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleDelete(ruleSet.id)}
                      disabled={ruleSetLoading[ruleSet.id]}
                      sx={{ p: 0.5 }}
                    >
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Stack>

                <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: 'block' }}>
                  最后更新：{new Date(ruleSet.updated_at).toLocaleString()}
                </Typography>

                {ruleSetLoading[ruleSet.tag] && <LinearProgress sx={{ mt: 1 }} />}
              </CardContent>
            </Card>
          </Grid>
        ))}

        {/* 规则卡片 */}
        {rules.map(rule => (
          <Grid item xs={12} sm={6} md={3} key={rule.id}>
            <Card variant="outlined" sx={{ ...innerCardStyle, height: '100%' }}>
              <CardContent sx={{ p: 1.5 }}>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                  <Box sx={{ width: 'calc(100% - 64px)' }}>
                    <Typography variant="subtitle2" noWrap>{rule.name}</Typography>
                    <Typography variant="caption" color="text.secondary" sx={{
                      display: '-webkit-box',
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: 'vertical',
                      overflow: 'hidden',
                      mb: 0.5
                    }}>
                      {rule.description}
                    </Typography>
                  </Box>
                  <Box>
                    <IconButton size="small" onClick={() => onEdit(rule)} sx={{ p: 0.5 }}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => onDelete(rule.id)} sx={{ p: 0.5 }}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Stack>

                <Stack direction="row" spacing={0.5}>
                  <Chip 
                    label={rule.outbound}
                    size="small"
                    color={
                      rule.outbound === '直连' ? 'success' :
                      rule.outbound === '拒绝' ? 'error' : 'warning'
                    }
                    sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                  />
                  {!rule.enabled && (
                    <Chip 
                      label="已禁用"
                      size="small"
                      color="default"
                      sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                    />
                  )}
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      {loading && <LinearProgress sx={{ mt: 2 }} />}

      {/* 错误提示 */}
      <Snackbar
        open={!!ruleSetError}
        autoHideDuration={3000}
        onClose={() => setRuleSetError(null)}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert onClose={() => setRuleSetError(null)} severity="error">
          {ruleSetError}
        </Alert>
      </Snackbar>
    </Box>
  );
};

// DNS 分流规则卡片
const DNSRulesCard = ({ rules, settings, loading, onAdd, onEdit, onDelete, onAdBlockToggle }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);

  return (
    <Box sx={{ mb: 3 }}>
      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
        <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
          DNS 分流规则
        </Typography>
        <Button
          variant="contained"
          size="small"
          startIcon={<AddIcon />}
          onClick={onAdd}
          sx={styles.actionButton}
        >
          添加规则
        </Button>
      </Stack>

      <Grid container spacing={2}>
        {/* 去广告规则卡片 */}
        <Grid item xs={12} sm={6} md={3}>
          <Card variant="outlined" sx={{ ...innerCardStyle, height: '100%' }}>
            <CardContent sx={{ p: 1.5 }}>
              <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                <Box sx={{ width: 'calc(100% - 64px)' }}>
                  <Typography variant="subtitle2" noWrap>去广告规则</Typography>
                  <Typography variant="caption" color="text.secondary">
                    内置去广告规则，自动拦截广告域名
                  </Typography>
                </Box>
                <FormControlLabel
                  control={
                    <Switch
                      size="small"
                      checked={settings?.enable_ad_block}
                      onChange={(e) => onAdBlockToggle(e.target.checked)}
                    />
                  }
                  label=""
                />
              </Stack>
              
              <Stack direction="row" spacing={0.5} sx={{ mt: 1 }}>
                <Chip 
                  label={settings?.enable_ad_block ? "已启用" : "已禁用"}
                  size="small"
                  color={settings?.enable_ad_block ? "success" : "default"}
                  sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                />
                <Chip 
                  label={`${settings?.ad_block_rule_count.toLocaleString()} 条规则`}
                  size="small"
                  sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                />
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        {/* 自定义规则卡片 */}
        {rules.map(rule => (
          <Grid item xs={12} sm={6} md={3} key={rule.id}>
            <Card variant="outlined" sx={{ ...innerCardStyle, height: '100%' }}>
              <CardContent sx={{ p: 1.5 }}>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                  <Box sx={{ width: 'calc(100% - 64px)' }}>
                    <Typography variant="subtitle2" noWrap>
                      {rule.type === 'domain' ? '域名' : 'IP'}
                    </Typography>
                    <Typography variant="caption" color="text.secondary" sx={{
                      display: '-webkit-box',
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: 'vertical',
                      overflow: 'hidden',
                      mb: 0.5
                    }}>
                      {rule.value}
                    </Typography>
                  </Box>
                  <Box>
                    <IconButton size="small" onClick={() => onEdit(rule)} sx={{ p: 0.5 }}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => onDelete(rule.id)} sx={{ p: 0.5 }}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Stack>

                <Stack direction="row" spacing={0.5}>
                  <Chip 
                    label={rule.action === 'direct' ? '国内' : '国外'}
                    size="small"
                    color={rule.action === 'direct' ? 'success' : 'warning'}
                    sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                  />
                  {!rule.enabled && (
                    <Chip 
                      label="已禁用"
                      size="small"
                      color="default"
                      sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
                    />
                  )}
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
};

// Hosts 文件管理卡片
const HostsCard = ({ hosts, loading, onAdd, onEdit, onDelete }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);

  return (
    <Card sx={cardStyle}>
      <CardContent>
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
          <Typography variant="h5" sx={{ fontWeight: 500, color: 'inherit' }}>
            Hosts 文件
          </Typography>
          <Button
            variant="contained"
            size="small"
            startIcon={<AddIcon />}
            onClick={onAdd}
            sx={styles.actionButton}
          >
            添加记录
          </Button>
        </Stack>

        {/* Hosts 记录列表 */}
        <Stack spacing={1}>
          {hosts.map(host => (
            <Card key={host.id} variant="outlined" sx={innerCardStyle}>
              <CardContent>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                  <Box>
                    <Typography variant="subtitle1">{host.domain}</Typography>
                    <Typography variant="body1">{host.ip}</Typography>
                    <Typography variant="body2" color="text.secondary">
                      {host.description}
                    </Typography>
                  </Box>
                  <Box>
                    <IconButton size="small" onClick={() => onEdit(host)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => onDelete(host.id)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Stack>
                
                {!host.enabled && (
                  <Chip 
                    label="已禁用"
                    size="small"
                    color="default"
                    sx={{ mt: 1 }}
                  />
                )}
              </CardContent>
            </Card>
          ))}
        </Stack>
      </CardContent>
    </Card>
  );
};

// RuleForm 组件
const RuleForm = ({ open, onClose, onSubmit, initialData = null }) => {
  const [formData, setFormData] = useState({
    name: '',
    type: 'domain', // domain, ip, mixed, remote
    domains: [],
    ips: [],
    urls: [], // 远程规则集 URLs
    format: 'mixed', // domain, ip, mixed
    outbound: '代理',
    description: '',
    enabled: true,
    priority: 0,
    ...initialData
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{initialData ? '编辑规则' : '添加规则'}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            label="规则名称"
            name="name"
            value={formData.name}
            onChange={handleChange}
            fullWidth
            required
          />

          <FormControl fullWidth>
            <InputLabel>规则类型</InputLabel>
            <Select
              name="type"
              value={formData.type}
              onChange={handleChange}
              label="规则类型"
            >
              <MenuItem value="domain">仅域名规则</MenuItem>
              <MenuItem value="ip">仅IP规则</MenuItem>
              <MenuItem value="mixed">混合规则（域名+IP）</MenuItem>
              <MenuItem value="remote">远程规则集</MenuItem>
            </Select>
          </FormControl>

          {(formData.type === 'domain' || formData.type === 'mixed') && (
          <TextField
              label="域名列表"
              name="domains"
              value={Array.isArray(formData.domains) ? formData.domains.join('\n') : ''}
              onChange={(e) => setFormData(prev => ({
                ...prev,
                domains: e.target.value.split('\n').filter(Boolean)
              }))}
            multiline
            rows={4}
              helperText="每行一个域名"
            fullWidth
            />
          )}

          {(formData.type === 'ip' || formData.type === 'mixed') && (
          <TextField
              label="IP列表"
              name="ips"
              value={Array.isArray(formData.ips) ? formData.ips.join('\n') : ''}
              onChange={(e) => setFormData(prev => ({
                ...prev,
                ips: e.target.value.split('\n').filter(Boolean)
              }))}
            multiline
              rows={4}
              helperText="每行一个IP或CIDR"
            fullWidth
          />
          )}

          {formData.type === 'remote' && (
            <>
          <TextField
                label="规则集URL列表"
                name="urls"
                value={Array.isArray(formData.urls) ? formData.urls.join('\n') : ''}
                onChange={(e) => setFormData(prev => ({
                  ...prev,
                  urls: e.target.value.split('\n').filter(Boolean)
                }))}
                multiline
                rows={4}
                required
                helperText={
                  "每行一个规则集URL，支持以下格式：\n" +
                  "1. @URL - 自动识别规则类型\n" +
                  "2. domain:URL - 域名规则\n" +
                  "3. ip:URL - IP规则"
                }
            fullWidth
          />
              <Alert severity="info" sx={{ mt: 1 }}>
                系统会根据URL前缀自动识别规则类型，如果没有指定前缀则自动识别
              </Alert>
            </>
          )}

          <FormControl fullWidth>
            <InputLabel>出站方式</InputLabel>
            <Select
              name="outbound"
              value={formData.outbound}
              onChange={handleChange}
              label="出站方式"
            >
              <MenuItem value="代理">代理</MenuItem>
              <MenuItem value="直连">直连</MenuItem>
              <MenuItem value="拒绝">拒绝</MenuItem>
            </Select>
          </FormControl>

          <TextField
            label="描述"
            name="description"
            value={formData.description}
            onChange={handleChange}
            fullWidth
            multiline
            rows={2}
          />

          <FormControlLabel
            control={
              <Switch
                checked={formData.enabled}
                onChange={(e) => setFormData(prev => ({
                  ...prev,
                  enabled: e.target.checked
                }))}
                name="enabled"
              />
            }
            label="启用规则"
          />

          <TextField
            label="优先级"
            name="priority"
            type="number"
            value={formData.priority}
            onChange={handleChange}
            fullWidth
            helperText="数字越大优先级越高"
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button onClick={handleSubmit} variant="contained">确定</Button>
      </DialogActions>
    </Dialog>
  );
};

// RuleSetCard 组件
const RuleSetCard = () => {
  const [loading, setLoading] = useState(false);
  const [ruleSets, setRuleSets] = useState([]);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  // 加载规则集信息
  const loadRuleSets = async () => {
    try {
      const data = await getRuleSets();
      setRuleSets(data);
    } catch (error) {
      setError('加载规则集信息失败');
    }
  };

  // 更新规则集
  const handleUpdate = async () => {
    setLoading(true);
    try {
      await updateRuleSets();
      setSuccess('规则集更新成功');
      loadRuleSets(); // 重新加载规则集信息
    } catch (error) {
      setError('规则集更新失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRuleSets();
  }, []);

  return (
    <Card variant="outlined" sx={{ ...innerCardStyle, height: '100%' }}>
      <CardContent sx={{ p: 1.5 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
          <Box sx={{ width: 'calc(100% - 64px)' }}>
            <Typography variant="subtitle2" noWrap>规则集管理</Typography>
            <Typography variant="caption" color="text.secondary">
              管理 GeoIP 和 GeoSite 规则集
            </Typography>
          </Box>
          <Box>
            <IconButton
              size="small"
              disabled={loading}
              onClick={handleUpdate}
              sx={{ p: 0.5 }}
            >
              <UpdateIcon fontSize="small" />
            </IconButton>
          </Box>
        </Stack>
        
        <Stack direction="row" spacing={0.5} sx={{ mt: 1 }}>
          {ruleSets.map((ruleSet) => (
            <Chip 
              key={ruleSet.tag}
              label={ruleSet.description}
              size="small"
              color="default"
              sx={{ height: 20, '& .MuiChip-label': { px: 1, fontSize: '0.75rem' } }}
            />
          ))}
        </Stack>

        {ruleSets.length > 0 && (
          <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: 'block' }}>
            最后更新：{new Date(ruleSets[0].updated_at).toLocaleString()}
          </Typography>
        )}

        {loading && <LinearProgress sx={{ mt: 1 }} />}
      </CardContent>

      {/* 错误和成功提示 */}
      <Snackbar
        open={!!error}
        autoHideDuration={3000}
        onClose={() => setError(null)}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert onClose={() => setError(null)} severity="error">
          {error}
        </Alert>
      </Snackbar>

      <Snackbar
        open={!!success}
        autoHideDuration={3000}
        onClose={() => setSuccess(null)}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert onClose={() => setSuccess(null)} severity="success">
          {success}
        </Alert>
      </Snackbar>
    </Card>
  );
};

// 主页面组件
const Rules = ({ mode, setMode, onDashboardClick }) => {
  const theme = useTheme();
  const styles = getCommonStyles(theme);

  const [rules, setRules] = useState([]);
  const [dnsRules, setDNSRules] = useState([]);
  const [dnsSettings, setDNSSettings] = useState(null);
  const [hosts, setHosts] = useState([]);
  const [nodeGroups, setNodeGroups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [ruleDialog, setRuleDialog] = useState(false);
  const [dnsRuleDialog, setDNSRuleDialog] = useState(false);
  const [hostDialog, setHostDialog] = useState(false);
  const [selectedRule, setSelectedRule] = useState(null);
  const [selectedDNSRule, setSelectedDNSRule] = useState(null);
  const [selectedHost, setSelectedHost] = useState(null);

  // 加载所有数据
  const fetchAllData = async () => {
    try {
      setLoading(true);
      const [
        rulesData,
        groupsData,
        dnsData,
        hostsData
      ] = await Promise.all([
        getRules(),
        getNodeGroups(),
        getDNSRules(),
        getHosts(),
      ]);
      
      setRules(rulesData);
      setNodeGroups(groupsData);
      setDNSRules(dnsData.rules);
      setDNSSettings(dnsData.settings);
      setHosts(hostsData);
    } catch (err) {
      setError('加载数据失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载
  useEffect(() => {
    fetchAllData();
  }, []);

  // 处理消息提示关闭
  const handleCloseSnackbar = () => {
    setError(null);
    setSuccess(null);
  };

  // 处理规则保存
  const handleSaveRule = async (rule) => {
    try {
      setLoading(true);
      
      // 处理远程规则集
      if (rule.type === 'remote') {
        // 确保 URLs 不为空
        if (!rule.urls || rule.urls.length === 0) {
          throw new Error('请至少添加一个规则集 URL');
        }

        // 处理每个 URL
        for (const url of rule.urls) {
          let format = 'mixed';  // 默认格式
          let actualUrl = url;

          // 解析 URL 格式
          if (url.startsWith('domain:')) {
            format = 'domain';
            actualUrl = url.substring(7);
          } else if (url.startsWith('ip:')) {
            format = 'ip';
            actualUrl = url.substring(3);
          } else if (url.startsWith('@')) {
            actualUrl = url.substring(1);
          }

          // 创建规则集
          const ruleSet = {
            name: rule.name,
            url: actualUrl,
            format: format,
            outbound: rule.outbound,
            description: rule.description,
            enabled: rule.enabled
          };

          await createRuleSet(ruleSet);
        }

        setSuccess('规则集创建成功');
      } else {
        // 处理普通规则
      if (rule.id) {
        await updateRule(rule.id, rule);
      } else {
        await createRule(rule);
      }
      setSuccess(rule.id ? '规则更新成功' : '规则创建成功');
      }

      await fetchAllData();
      setRuleDialog(false);
      setSelectedRule(null);
    } catch (err) {
      setError(err.message || '保存规则失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 处理规则删除
  const handleDeleteRule = async (id) => {
    try {
      setLoading(true);
      await deleteRule(id);
      await fetchAllData();
      setSuccess('规则删除成功');
    } catch (err) {
      setError('删除规则失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 处理 DNS 规则保存
  const handleSaveDNSRule = async (rule) => {
    try {
      setLoading(true);
      if (rule.id) {
        await updateDNSRule(rule.id, rule);
      } else {
        await createDNSRule(rule);
      }
      await fetchAllData();
      setSuccess(rule.id ? 'DNS 规则更新成功' : 'DNS 规则创建成功');
      setDNSRuleDialog(false);
      setSelectedDNSRule(null);
    } catch (err) {
      setError('保存 DNS 规则失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 处理 DNS 规则删除
  const handleDeleteDNSRule = async (id) => {
    try {
      setLoading(true);
      await deleteDNSRule(id);
      await fetchAllData();
      setSuccess('DNS 规则删除成功');
    } catch (err) {
      setError('删除 DNS 规则失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 处理 Hosts 记录保存
  const handleSaveHost = async (host) => {
    try {
      setLoading(true);
      if (host.id) {
        await updateHost(host.id, host);
      } else {
        await createHost(host);
      }
      await fetchAllData();
      setSuccess(host.id ? 'Hosts 记录更新成功' : 'Hosts 记录创建成功');
      setHostDialog(false);
      setSelectedHost(null);
    } catch (err) {
      setError('保存 Hosts 记录失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 处理 Hosts 记录删除
  const handleDeleteHost = async (id) => {
    try {
      setLoading(true);
      await deleteHost(id);
      await fetchAllData();
      setSuccess('Hosts 记录删除成功');
    } catch (err) {
      setError('删除 Hosts 记录失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 处理去广告规则开关
  const handleAdBlockToggle = async (checked) => {
    try {
      setLoading(true);
      await updateDNSSettings({
        ...dnsSettings,
        enable_ad_block: checked,
      });
      await fetchAllData();
      setSuccess('设置更新成功');
    } catch (err) {
      setError('更新设置失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader 
        title="规则管理" 
        mode={mode} 
        setMode={setMode} 
        onDashboardClick={onDashboardClick}
        styles={styles}
      />
      
      {loading && <LinearProgress sx={{ mb: 2 }} />}

      <Stack spacing={3}>
      {/* SingBox 路由规则 */}
      <SingBoxRules
        rules={rules}
        loading={loading}
        onAdd={() => setRuleDialog(true)}
        onEdit={rule => {
          setSelectedRule(rule);
          setRuleDialog(true);
        }}
        onDelete={handleDeleteRule}
        nodeGroups={nodeGroups}
      />

      {/* DNS 分流规则 */}
      <DNSRulesCard
        rules={dnsRules}
        settings={dnsSettings}
        loading={loading}
        onAdd={() => setDNSRuleDialog(true)}
        onEdit={rule => {
          setSelectedDNSRule(rule);
          setDNSRuleDialog(true);
        }}
        onDelete={handleDeleteDNSRule}
        onAdBlockToggle={handleAdBlockToggle}
      />

      {/* Hosts 文件管理 */}
      <HostsCard
        hosts={hosts}
        loading={loading}
        onAdd={() => setHostDialog(true)}
        onEdit={host => {
          setSelectedHost(host);
          setHostDialog(true);
        }}
        onDelete={handleDeleteHost}
      />
      </Stack>

      {/* 规则编辑对话框 */}
      <RuleForm
        open={ruleDialog}
        onClose={() => {
          setRuleDialog(false);
          setSelectedRule(null);
        }}
        onSubmit={handleSaveRule}
        initialData={selectedRule}
      />

      {/* DNS 规则编辑对话框 */}
      <RuleForm
        open={dnsRuleDialog}
        onClose={() => {
          setDNSRuleDialog(false);
          setSelectedDNSRule(null);
        }}
        onSubmit={handleSaveDNSRule}
        initialData={selectedDNSRule}
      />

      {/* Hosts 记录编辑对话框 */}
      <RuleForm
        open={hostDialog}
        onClose={() => {
          setHostDialog(false);
          setSelectedHost(null);
        }}
        onSubmit={handleSaveHost}
        initialData={selectedHost}
      />

      {/* 消息提示 */}
      <Snackbar
        open={!!error || !!success}
        autoHideDuration={3000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert
          onClose={handleCloseSnackbar}
          severity={error ? 'error' : 'success'}
          sx={{ width: '100%' }}
        >
          {error || success}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default Rules; 