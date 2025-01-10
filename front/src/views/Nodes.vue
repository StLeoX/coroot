<template>
    <div>
        <v-progress-linear indeterminate v-if="loading" color="green" />

        <v-alert v-if="error" color="red" icon="mdi-alert-octagon-outline" outlined text>
            {{ error }}
        </v-alert>

        <Table v-if="nodes && nodes.rows" :header="nodes.header" :rows="nodes.rows" />
        <NoData v-else-if="!loading && !error" />
        
    </div>
</template>

<script>
import NoData from '@/components/NoData.vue';
import Table from '@/components/Table.vue';

export default {
    components: { Table, NoData },

    data() {
        return {
            nodes: null,
            loading: false,
            error: '',
        };
    },

    mounted() {
        this.get();
        this.$events.watch(this, this.get, 'refresh');
    },

    methods: {
        get() {
            this.loading = true;
            this.error = '';
            this.$api.getOverview('nodes', '', (data, error) => {
                this.loading = false;
                if (error) {
                    this.error = error;
                    return;
                }
                this.nodes = data.nodes;
            });
        },
    },
};
</script>
