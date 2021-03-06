<template>
    <div class="table" v-bkloading="{ isLoading: $loading(propertyRequest, instanceRequest) }">
        <div class="table-info clearfix">
            <div class="info-title fl" @click="localVisible = !localVisible">
                <i class="icon bk-icon icon-right-shape"
                    :class="{ 'is-open': localVisible }">
                </i>
                <span class="title-text">{{title}}</span>
                <span class="title-count">({{instances.length}})</span>
            </div>
            <div class="info-pagination fr" v-show="pagination.count">
                <span class="pagination-info">{{getPaginationInfo()}}</span>
                <span class="pagination-toggle">
                    <i class="pagination-icon bk-icon icon-angle-left"
                        :class="{ disabled: pagination.current === 1 }"
                        @click="togglePage(-1)">
                    </i>
                    <i class="pagination-icon bk-icon icon-angle-right"
                        :class="{ disabled: pagination.current === totalPage }"
                        @click="togglePage(1)">
                    </i>
                </span>
            </div>
        </div>
        <cmdb-collapse-transition>
            <cmdb-table class="association-table"
                v-show="localVisible"
                :header="header"
                :list="flatternList"
                :show-footer="false"
                :sortable="false"
                :max-height="462"
                :empty-height="40">
                <template slot="__operation__" slot-scope="{ item }">
                    <span class="text-primary" @click="showTips($event, item)">
                        {{$t('Association["取消关联"]')}}
                    </span>
                </template>
            </cmdb-table>
        </cmdb-collapse-transition>
        <div class="confirm-tips" ref="confirmTips" v-click-outside="hideTips" v-show="confirm.item">
            <p class="tips-content">{{$t('Association["确认取消"]')}}</p>
            <div class="tips-option">
                <bk-button class="tips-button" type="primary" @click="cancelAssociation">{{$t('Common["确认"]')}}</bk-button>
                <bk-button class="tips-button" type="default" @click="hideTips">{{$t('Common["取消"]')}}</bk-button>
            </div>
        </div>
    </div>
</template>

<script>
    import { mapGetters } from 'vuex'
    import { OPERATION, RESOURCE_HOST } from '../router.config.js'
    export default {
        name: 'cmdb-host-association-list-table',
        props: {
            type: {
                type: String,
                required: true
            },
            id: {
                type: String,
                required: true
            },
            associationType: {
                type: Object,
                required: true
            },
            visible: Boolean
        },
        data () {
            return {
                properties: [],
                list: [],
                pagination: {
                    count: 0,
                    current: 1,
                    size: 10
                },
                localVisible: this.visible,
                confirm: {
                    instance: null,
                    item: null,
                    target: null
                }
            }
        },
        computed: {
            ...mapGetters('hostDetails', [
                'sourceInstances',
                'targetInstances'
            ]),
            expandAll () {
                return this.$store.state.hostDetails.expandAll
            },
            indeterminate () {
                return this.$store.state.hostDetails.indeterminate
            },
            updateAuth () {
                const isResourceHost = this.$route.name === RESOURCE_HOST
                if (isResourceHost) {
                    return this.$isAuthorized(OPERATION.U_RESOURCE_HOST)
                }
                return this.$isAuthorized(OPERATION.U_HOST)
            },
            flatternList () {
                return this.$tools.flatternList(this.properties, this.list)
            },
            hostId () {
                return parseInt(this.$route.params.id)
            },
            model () {
                return this.$store.getters['objectModelClassify/getModelById'](this.id)
            },
            isBusinessModel () {
                return !!this.$tools.getMetadataBiz(this.model)
            },
            title () {
                const desc = this.type === 'source' ? this.associationType.src_des : this.associationType.dest_des
                return `${desc}-${this.model.bk_obj_name}`
            },
            propertyRequest () {
                return `get_${this.id}_association_list_table_properties`
            },
            instanceRequest () {
                return `get_${this.id}_association_list_table_instances`
            },
            page () {
                return {
                    limit: this.pagination.size,
                    start: (this.pagination.current - 1) * this.pagination.size
                }
            },
            totalPage () {
                return Math.ceil(this.pagination.count / this.pagination.size)
            },
            instances () {
                const topology = this.type === 'source' ? this.targetInstances : this.sourceInstances
                const data = topology.find(data => data.bk_obj_id === this.id) || {}
                return data.children || []
            },
            instanceIds () {
                return this.instances.map(instance => instance.bk_inst_id)
            },
            header () {
                const headerProperties = this.$tools.getDefaultHeaderProperties(this.properties)
                const header = headerProperties.map(property => {
                    return {
                        id: property.bk_property_id,
                        name: property.bk_property_name
                    }
                })
                if (this.updateAuth) {
                    header.push({
                        id: '__operation__',
                        name: this.$t('Common["操作"]'),
                        width: 150
                    })
                }
                return header
            }
        },
        watch: {
            visible (visible) {
                this.localVisible = visible
            },
            localVisible (visible) {
                if (visible) {
                    this.getData()
                }
                this.$store.commit('hostDetails/updateExpandCount', visible ? 1 : -1)
                this.checkExpandAllState()
            },
            instances () {
                if (this.localVisible) {
                    this.getData()
                }
            },
            expandAll (expandAll) {
                if (!this.indeterminate) {
                    this.localVisible = expandAll
                }
            }
        },
        created () {
            this.$store.commit('hostDetails/updateTableCount', 1)
            if (this.visible) {
                this.getData()
                this.$store.commit('hostDetails/updateExpandCount', 1)
            }
        },
        beforeDestroy () {
            this.$store.commit('hostDetails/updateTableCount', -1)
        },
        methods: {
            getData () {
                this.getProperties()
                this.getInstances()
            },
            async getProperties () {
                try {
                    this.properties = await this.$store.dispatch('objectModelProperty/searchObjectAttribute', {
                        params: this.$injectMetadata({
                            bk_obj_id: this.id
                        }, {
                            inject: this.isBusinessModel
                        }),
                        config: {
                            fromCache: true,
                            requestId: this.propertyRequest
                        }
                    })
                } catch (e) {
                    console.error(e)
                    this.properties = []
                }
            },
            async getInstances () {
                let promise
                const config = {
                    requestId: this.instanceRequest,
                    cancelPrevious: true,
                    globalError: false
                }
                try {
                    switch (this.id) {
                        case 'host':
                            promise = this.getHostInstances(config)
                            break
                        case 'biz':
                            promise = this.getBusinessInstances(config)
                            break
                        default:
                            promise = this.getModelInstances(config)
                    }
                    const data = await promise
                    this.list = data.info
                    this.pagination.count = data.count
                } catch (e) {
                    console.error(e)
                    this.list = []
                    this.pagination.count = 0
                }
            },
            getHostInstances (config) {
                const models = ['biz', 'set', 'module', 'host']
                const hostCondition = {
                    field: 'bk_host_id',
                    operator: '$in',
                    value: this.instanceIds
                }
                const condition = models.map(model => {
                    return {
                        bk_obj_id: model,
                        fields: [],
                        condition: model === 'host' ? [hostCondition] : []
                    }
                })
                return this.$store.dispatch('hostSearch/searchHost', {
                    params: this.$injectMetadata({
                        bk_biz_id: -1,
                        condition,
                        id: {
                            data: [],
                            exact: 0,
                            flag: 'bk_host_innerip|bk_host_outerip'
                        },
                        page: {
                            ...this.page,
                            sort: 'bk_host_id'
                        }
                    }),
                    config
                }).then(data => {
                    return {
                        count: data.count,
                        info: data.info.map(item => item.host)
                    }
                })
            },
            getBusinessInstances (config) {
                return this.$store.dispatch('objectBiz/searchBusiness', {
                    params: {
                        condition: {
                            bk_biz_id: {
                                $in: this.instanceIds
                            }
                        },
                        fields: [],
                        page: {
                            ...this.page,
                            sort: 'bk_biz_id'
                        }
                    },
                    config
                })
            },
            getModelInstances (config) {
                return this.$store.dispatch('objectCommonInst/searchInst', {
                    objId: this.id,
                    params: {
                        fields: {},
                        condition: {
                            [this.id]: [{
                                field: 'bk_inst_id',
                                operator: '$in',
                                value: this.instanceIds
                            }]
                        },
                        page: {
                            ...this.page,
                            sort: 'bk_inst_id'
                        }
                    },
                    config
                }).then(data => {
                    data = data || {
                        count: 0,
                        info: []
                    }
                    return data
                })
            },
            async cancelAssociation () {
                const item = this.confirm.item
                const keyMap = {
                    host: 'bk_host_id',
                    biz: 'bk_biz_id'
                }
                const idKey = keyMap[this.id] || 'bk_inst_id'
                try {
                    const associationInstance = this.instances.find(instance => instance.bk_inst_id === item[idKey])
                    await this.$store.dispatch('objectAssociation/deleteInstAssociation', {
                        id: associationInstance.asso_id,
                        config: {
                            data: this.$injectMetadata({}, {
                                inject: !!this.$tools.getMetadataBiz(this.model)
                            })
                        }
                    })
                    this.$store.commit('hostDetails/deleteAssociation', {
                        type: this.type,
                        model: this.id,
                        association: associationInstance
                    })
                    this.$nextTick(() => {
                        this.pagination.current = 1
                        this.getInstances()
                    })
                    this.$success(this.$t('Association["取消关联成功"]'))
                } catch (e) {
                    console.error(e)
                }
            },
            getPaginationInfo () {
                return this.$tc('Common["页码"]', this.pagination.current, {
                    current: this.pagination.current,
                    count: this.pagination.count,
                    total: this.totalPage
                })
            },
            togglePage (step) {
                const current = this.pagination.current
                const newCurrent = current + step
                if (newCurrent < 1 || newCurrent > this.totalPage) {
                    return false
                }
                this.pagination.current = newCurrent
                this.getInstances()
            },
            hideTips (event) {
                if (event && event.target === this.confirm.target) {
                    return false
                }
                this.confirm.instance && this.confirm.instance.setVisible(false)
            },
            showTips (event, item) {
                this.confirm.item = item
                this.confirm.target = event.target
                this.confirm.instance && this.confirm.instance.destroy()
                this.confirm.instance = this.$tooltips({
                    duration: 100,
                    theme: 'light',
                    zIndex: 9999,
                    width: 200,
                    container: document.body,
                    target: event.target
                })
                this.confirm.instance.$el.append(this.$refs.confirmTips)
            },
            checkExpandAllState () {
                this.$nextTick(() => {
                    const state = this.$store.state.hostDetails
                    const tableCount = state.tableCount
                    const expandCount = state.expandCount
                    this.$store.commit('hostDetails/setExpandIndeterminate', expandCount !== 0 && tableCount !== expandCount)
                    this.$store.commit('hostDetails/toggleExpandAll', tableCount === expandCount)
                })
            }
        }
    }
</script>

<style lang="scss" scoped>
    .table {
        margin: 0 0 12px 0;
    }
    .table-info {
        height: 42px;
        padding: 0 20px;
        border-radius: 2px 2px 0 0;
        line-height: 42px;
        background-color: #DCDEE5;
        font-size: 14px;
    }
    .info-title {
        cursor: pointer;
        .icon {
            display: inline-block;
            vertical-align: middle;
            transition: transform .2s linear;
            &.is-open {
                transform: rotate(90deg);
            }
        }
        .title-text {
            color: #000;
        }
        .title-count {
            color: #8b8d95;
        }
    }
    .info-pagination {
        color: #8b8d95;
        .pagination-toggle {
            .pagination-icon {
                transform: scale(.5);
                font-size: 20px;
                color: #979BA5;
                cursor: pointer;
                &.disabled {
                    color: #C4C6CC;
                    cursor: not-allowed;
                }
            }
        }
    }
    .confirm-tips {
        padding: 9px;
        text-align: center;
        .tips-content {
            color: $cmdbTextColor;
            line-height: 20px;
        }
        .tips-option {
            margin: 12px 0 0 0;
            .tips-button {
                height: 26px;
                line-height: 24px;
                padding: 0 16px;
                min-width: 56px;
                font-size: 12px;
            }
        }
    }
</style>
