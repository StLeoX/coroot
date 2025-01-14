<template>
    <div>
        <div class="my-4">
            <v-tabs :value="view" height="40" show-arrows slider-size="2">
                <template v-for="(name, view) in views">
                    <v-tab
                        v-if="name"
                        :to="{
                            params: { view, id: undefined, report: undefined },
                            query: $utils.contextQuery(),
                        }"
                        :tab-value="view"
                    >
                        {{ name }}
                    </v-tab>
                </template>
            </v-tabs>
        </div>

        <template v-if="view === 'applications'">
            <Application v-if="id" :id="id" :report="report" />
            <Applications v-else />
        </template>

        <template v-if="view === 'map'">
            <ServiceMap />
        </template>

        <template v-if="view === 'nodes'">
            <Node v-if="id" :name="id" />
            <Nodes v-else />
        </template>

        <template v-if="view === 'traces'">
            <Traces />
        </template>
    </div>
</template>

<script>
import Applications from '@/views/Applications.vue';
import Application from '@/views/Application.vue';
import ServiceMap from '@/views/ServiceMap.vue';
import Traces from '@/views/Traces.vue';
import Nodes from '@/views/Nodes.vue';
import Node from '@/views/Node.vue';

export default {
    components: {
        Applications,
        Application,
        ServiceMap,
        Traces,
        Nodes,
        Node,
    },
    props: {
        view: String,
        id: String,
        report: String,
    },

    computed: {
        views() {
            return {
                applications: 'Containers',
                map: 'Service Map',
                traces: 'Traces',
                nodes: 'Nodes',
            };
        },
    },

    watch: {
        view: {
            handler(v) {
                if (!this.views[v]) {
                    this.$router.replace({ params: { view: 'applications' } }).catch((err) => err);
                }
            },
            immediate: true,
        },
    },
};
</script>

<style scoped></style>
