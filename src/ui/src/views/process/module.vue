<template>
    <div class="module-bind-wrapper">
        <cmdb-table class="module-bind-table" ref="table"
            :loading="$loading('getModuleList')"
            :header="table.header"
            :list="table.list"
            :wrapper-minus-height="150"
            :sortable="false">
            <template slot="is_bind" slot-scope="{ item }">
                <bk-button type="primary"
                    v-if="item['is_bind'] === 1"
                    :loading="$loading(`${item['bk_module_name']}Bind`)"
                    :disabled="!unbindModuleAuth"
                    @click="unbindModule(item)">
                    {{$t("ProcessManagement['已绑定']")}}
                </bk-button>
                <bk-button type="default"
                    v-else
                    :loading="$loading(`${item['bk_module_name']}Bind`)"
                    :disabled="!bindModuleAuth"
                    @click="bindModule(item)">
                    {{$t("ProcessManagement['未绑定']")}}
                </bk-button>
            </template>
        </cmdb-table>
    </div>
</template>

<script>
    import { mapActions } from 'vuex'
    import { OPERATION } from './router.config.js'
    export default {
        props: {
            processId: {
                type: [String, Number],
                required: true
            },
            bizId: {
                type: [String, Number],
                required: true
            }
        },
        data () {
            return {
                table: {
                    header: [{
                        id: 'bk_module_name',
                        name: this.$t("ProcessManagement['模块名']")
                    }, {
                        id: 'set_num',
                        name: this.$t("ProcessManagement['所属集群数']")
                    }, {
                        id: 'is_bind',
                        name: this.$t("ProcessManagement['状态']")
                    }],
                    list: [],
                    maxHeight: 0
                }
            }
        },
        computed: {
            bindModuleAuth () {
                return this.$isAuthorized(OPERATION.PROCESS_BIND_MODULE)
            },
            unbindModuleAuth () {
                return this.$isAuthorized(OPERATION.PROCESS_UNBIND_MODULE)
            }
        },
        created () {
            this.getModuleList()
        },
        methods: {
            ...mapActions('procConfig', [
                'bindProcessModule',
                'deleteProcessModuleBinding',
                'getProcessBindModule'
            ]),
            bindModule (item) {
                const moduleName = item['bk_module_name'].replace(' ', '')
                this.bindProcessModule({
                    bizId: this.bizId,
                    processId: this.processId,
                    moduleName,
                    config: {
                        requestId: `${item['bk_module_name']}Bind`
                    }
                })
                item['is_bind'] = 1
            },
            unbindModule (item) {
                const moduleName = item['bk_module_name'].replace(' ', '')
                this.deleteProcessModuleBinding({
                    bizId: this.bizId,
                    processId: this.processId,
                    moduleName,
                    config: {
                        requestId: `${item['bk_module_name']}Bind`
                    }
                })
                item['is_bind'] = 0
            },
            async getModuleList () {
                const res = await this.getProcessBindModule({
                    bizId: this.bizId,
                    processId: this.processId,
                    config: {
                        requestId: 'getModuleList'
                    }
                })
                this.table.list = this.sortModule(res)
                this.calcMaxHeight()
            },
            sortModule (data) {
                const bindedModule = []
                const unbindModule = []
                data.forEach(module => {
                    module['is_bind'] ? bindedModule.push(module) : unbindModule.push(module)
                })
                bindedModule.sort((moduleA, moduleB) => {
                    return moduleA['bk_module_name'].localeCompare(moduleB['bk_module_name'])
                })
                unbindModule.sort((moduleA, moduleB) => {
                    return moduleA['bk_module_name'].localeCompare(moduleB['bk_module_name'])
                })
                return [...bindedModule, ...unbindModule]
            },
            calcMaxHeight () {
                this.table.maxHeight = document.body.getBoundingClientRect().height - 160
            }
        }
    }
</script>

<style lang="scss" scoped>
    .module-bind-wrapper {
        padding-top: 20px;
        .bk-button {
            padding: 1px 7px 2px;
            height: 22px;
            line-height: 20px;
            font-size: 12px;
        }
    }
</style>
