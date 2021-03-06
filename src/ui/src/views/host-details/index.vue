<template>
    <div class="details-layout">
        <cmdb-host-info
            ref="info"
            @info-toggle="setInfoHeight">
        </cmdb-host-info>
        <bk-tab class="details-tab"
            :active-name.sync="active"
            :style="{
                '--infoHeight': infoHeight
            }">
            <bk-tabpanel name="property" :title="$t('HostDetails[\'主机属性\']')">
                <cmdb-host-property></cmdb-host-property>
            </bk-tabpanel>
            <bk-tabpanel name="association" :title="$t('HostDetails[\'关联\']')">
                <cmdb-host-association v-if="active === 'association'"></cmdb-host-association>
            </bk-tabpanel>
            <bk-tabpanel name="status" :title="$t('HostDetails[\'实时状态\']')">
                <cmdb-host-status v-if="active === 'status'"></cmdb-host-status>
            </bk-tabpanel>
            <bk-tabpanel name="history" :title="$t('HostDetails[\'变更记录\']')">
                <cmdb-host-history v-if="active === 'history'"></cmdb-host-history>
            </bk-tabpanel>
        </bk-tab>
    </div>
</template>

<script>
    import { mapState } from 'vuex'
    import cmdbHostInfo from './children/info.vue'
    import cmdbHostAssociation from './children/association.vue'
    import cmdbHostProperty from './children/property.vue'
    import cmdbHostStatus from './children/status.vue'
    import cmdbHostHistory from './children/history.vue'
    export default {
        components: {
            cmdbHostInfo,
            cmdbHostAssociation,
            cmdbHostProperty,
            cmdbHostStatus,
            cmdbHostHistory
        },
        data () {
            return {
                active: 'property',
                infoHeight: '81px'
            }
        },
        computed: {
            ...mapState('hostDetails', ['info']),
            id () {
                return parseInt(this.$route.params.id)
            },
            business () {
                const business = parseInt(this.$route.params.business)
                if (isNaN(business)) {
                    return -1
                }
                return business
            }
        },
        watch: {
            info (info) {
                this.$store.commit('setHeaderTitle', `${this.$t('HostDetails["主机详情"]')}(${info.host.bk_host_innerip})`)
            },
            id () {
                this.getData()
            },
            business () {
                this.getData()
            },
            active (active) {
                if (active !== 'association') {
                    this.$store.commit('hostDetails/toggleExpandAll', false)
                    this.$store.commit('hostDetails/setExpandIndeterminate', true)
                }
            }
        },
        created () {
            this.$store.commit('setHeaderTitle', this.$t('HostDetails["主机详情"]'))
            this.$store.commit('setHeaderStatus', { back: true })
            this.getData()
        },
        methods: {
            getData () {
                this.getProperties()
                this.getPropertyGroups()
                this.getHostInfo()
            },
            async getHostInfo () {
                try {
                    const { info } = await this.$store.dispatch('hostSearch/searchHost', {
                        params: this.getSearchHostParams()
                    })
                    if (info.length) {
                        this.$store.commit('hostDetails/setHostInfo', info[0])
                    } else {
                        this.$router.replace({ name: 404 })
                    }
                } catch (e) {
                    console.error(e)
                    this.$store.commit('hostDetails/setHostInfo', null)
                }
            },
            getSearchHostParams () {
                const hostCondition = {
                    field: 'bk_host_id',
                    operator: '$eq',
                    value: this.id
                }
                return this.$injectMetadata({
                    bk_biz_id: this.business,
                    condition: ['biz', 'set', 'module', 'host'].map(model => {
                        return {
                            bk_obj_id: model,
                            condition: model === 'host' ? [hostCondition] : [],
                            fields: []
                        }
                    }),
                    ip: { flag: 'bk_host_innerip', exact: 1, data: [] }
                })
            },
            async getProperties () {
                try {
                    const properties = await this.$store.dispatch('objectModelProperty/searchObjectAttribute', {
                        params: this.$injectMetadata({
                            bk_obj_id: 'host'
                        })
                    })
                    this.$store.commit('hostDetails/setHostProperties', properties)
                } catch (e) {
                    console.error(e)
                    this.$store.commit('hostDetails/setHostProperties', [])
                }
            },
            async getPropertyGroups () {
                try {
                    const propertyGroups = await this.$store.dispatch('objectModelFieldGroup/searchGroup', {
                        objId: 'host',
                        params: this.$injectMetadata()
                    })
                    this.$store.commit('hostDetails/setHostPropertyGroups', propertyGroups)
                } catch (e) {
                    console.error(e)
                    this.$store.commit('hostDetails/setHostPropertyGroups', [])
                }
            },
            setInfoHeight (height) {
                this.infoHeight = height + 'px'
            }
        }
    }
</script>

<style lang="scss" scoped>
    .details-layout {
        padding: 0;
        height: 100%;
        .details-tab {
            height: calc(100% - var(--infoHeight)) !important;
            min-height: 400px;
        }
    }
</style>
