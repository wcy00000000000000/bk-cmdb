<template>
    <div class="history-wrapper">
        <div class="filter-wrapper clearfix">
            <cmdb-form-date-range class="date-range" v-model="filter['last_time']"></cmdb-form-date-range>
            <bk-selector
                class="selector"
                :placeholder="$t('NetworkDiscovery[\'全部变更\']')"
                :allow-clear="true"
                :list="changeList"
                :selected.sync="filter['action']"
            ></bk-selector>
            <bk-selector
                class="selector"
                :placeholder="$t('NetworkDiscovery[\'全部类型\']')"
                :allow-clear="true"
                :list="typeList"
                :selected.sync="filter['bk_obj_id']"
            ></bk-selector>
            <input type="text" class="cmdb-form-input" :placeholder="$t('NetworkDiscovery[\'请输入云区域名称\']')" v-model.trim="filter['bk_cloud_name']">
            <input type="text" class="cmdb-form-input" :placeholder="$t('NetworkDiscovery[\'请输入IP\']')" v-model.trim="filter['bk_host_innerip']">
            <bk-button type="primary" @click="getTableData">
                {{$t("Common['查询']")}}
            </bk-button>
        </div>
        <cmdb-table
            class="history-table"
            :loading="$loading('searchNetcollect')"
            :header="table.header"
            :list="table.list"
            :pagination.sync="table.pagination"
            :default-sort="table.defaultSort"
            @handleSortChange="handleSortChange"
            @handleSizeChange="handleSizeChange"
            @handlePageChange="handlePageChange">
            <template slot="action" slot-scope="{ item }">
                <span :class="{ 'color-danger': item.action === 'delete', 'color-warning': item.action === 'update' }">{{actionMap[item.action]}}</span>
            </template>
            <template slot="last_time" slot-scope="{ item }">
                {{$tools.formatTime(item['last_time'], 'YYYY-MM-DD')}}
            </template>
            <template slot="success" slot-scope="{ item }">
                <span :class="item.success ? 'color-success' : 'color-danger'">{{item.success ? $t("Inst['成功']") : $t("EventPush['失败']")}}</span>
            </template>
        </cmdb-table>
    </div>
</template>

<script>
    import { mapActions } from 'vuex'
    export default {
        data () {
            return {
                filter: {
                    last_time: [],
                    action: '',
                    bk_obj_id: '',
                    bk_cloud_name: '',
                    bk_host_innerip: ''
                },
                changeList: [{
                    id: 'create',
                    name: this.$t("Common['新增']")
                }, {
                    id: 'update',
                    name: this.$t("NetworkDiscovery['变更']")
                }, {
                    id: 'delete',
                    name: this.$t("Common['删除']")
                }],
                typeList: [{
                    id: 'switch',
                    name: this.$t("NetworkDiscovery['交换机']")
                }, {
                    id: 'host',
                    name: this.$t("Hosts['主机']")
                }],
                table: {
                    header: [{
                        id: 'action',
                        name: this.$t('NetworkDiscovery["变更方式"]')
                    }, {
                        id: 'bk_cloud_name',
                        name: this.$t('Hosts["云区域"]')
                    }, {
                        id: 'bk_obj_name',
                        name: this.$t('ModelManagement["类型"]')
                    }, {
                        id: 'bk_inst_key',
                        name: this.$t('NetworkDiscovery["唯一标识"]')
                    }, {
                        id: 'bk_host_innerip',
                        name: 'IP'
                    }, {
                        id: 'configuration',
                        name: this.$t('NetworkDiscovery["配置信息"]')
                    }, {
                        id: 'last_time',
                        name: this.$t('NetworkDiscovery["发现时间"]')
                    }, {
                        id: 'success',
                        name: this.$t('ProcessManagement["状态"]')
                    }],
                    list: [],
                    pagination: {
                        count: 0,
                        size: 10,
                        current: 1
                    },
                    defaultSort: '-last_time',
                    sort: '-last_time'
                },
                actionMap: {
                    'create': this.$t("Common['新增']"),
                    'update': this.$t("NetworkDiscovery['变更']"),
                    'delete': this.$t("Common['删除']")
                }
            }
        },
        computed: {
            params () {
                const params = {
                    bk_cloud_name: this.filter['bk_cloud_name'],
                    bk_host_innerip: this.filter['bk_host_innerip'],
                    bk_obj_id: this.filter['bk_obj_id'],
                    action: this.filter['action'],
                    last_time: this.filter['last_time']
                }
                return params
            }
        },
        created () {
            this.$route.meta.title = this.$t('NetworkDiscovery["完成历史"]')
            this.getTableData()
        },
        methods: {
            ...mapActions('netDiscovery', [
                'searchNetcollectHistory'
            ]),
            async getTableData () {
                const res = await this.searchNetcollectHistory({ params: this.params, config: { requestId: 'searchNetcollect' } })
                this.table.pagination.count = res.count
                this.table.list = res.info
            },
            handleSortChange (sort) {
                this.table.sort = sort
                this.handlePageChange(1)
            },
            handleSizeChange (size) {
                this.table.pagination.size = size
                this.handlePageChange(1)
            },
            handlePageChange (page) {
                this.table.pagination.current = page
                this.getTableData()
            }
        }
    }
</script>

<style lang="scss" scoped>
    .filter-wrapper {
        .date-range {
            float: left;
            margin-right: 10px;
            width: calc((100% - 135px) * (240 / 920));
        }
        .selector {
            float: left;
            margin-right: 10px;
            width: calc((100% - 135px) * (140 / 920));
        }
        .cmdb-form-input {
            float: left;
            margin-right: 10px;
            width: calc((100% - 135px) * (200 / 920));
        }
        .bk-button {
            width: 85px;
        }
    }
    .history-table {
        margin-top: 20px;
    }
</style>
