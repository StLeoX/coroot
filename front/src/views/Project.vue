<template>
    <div class="mx-auto">
        <h1 class="text-h5 my-5">Configuration</h1>

        <v-tabs height="40" show-arrows slider-size="2">
            <v-tab v-for="t in tabs" :key="t.id" :to="{ params: { tab: t.id } }" :disabled="t.disabled" exact>
                {{ t.name }}
            </v-tab>
        </v-tabs>

        <template v-if="!tab">
            <h2 class="text-h5 my-5">Project name</h2>
            <ProjectSettings :projectId="projectId" />

            <template v-if="projectId">
                <h2 class="text-h5 mt-10 mb-5">Status</h2>
                <ProjectStatus :projectId="projectId" />

                <h2 class="text-h5 mt-10 mb-5">Danger zone</h2>
                <ProjectDelete :projectId="projectId" />
            </template>
        </template>

        <template v-if="tab === 'prometheus'">
            <h1 class="text-h5 my-5">
                Prometheus integration
            </h1>
            <IntegrationPrometheus />
        </template>

        <template v-if="tab === 'clickhouse'">
            <h1 class="text-h5 my-5">
                ClickHouse integration
            </h1>
            <p>
            </p>
            <IntegrationClickhouse />
        </template>


        <template v-if="tab === 'inspections'">
            <h1 class="text-h5 my-5">
                Inspection configs
            </h1>
            <Inspections />
        </template>

        <template v-if="tab === 'applications'">
            <h2 class="text-h5 my-5" id="categories">
                Application categories
            </h2>
            <ApplicationCategories />

            <h2 class="text-h5 mt-10 mb-5" id="custom-applications">
                Custom applications
            </h2>


            <ul>
                <li><b>Kubernetes metadata</b>: Pods are grouped into Deployments, StatefulSets, etc.</li>
                <li>
                    <b>Non-Kubernetes containers</b>: Containers such as Docker containers or Systemd units are grouped into applications by their
                    names. For example, Systemd services named <var>mysql</var> on different hosts are grouped into a single application called
                    <var>mysql</var>.
                </li>
            </ul>

            <!-- <p class="my-5">
                This default approach works well in most cases. However, since no one knows your system better than you do, Coroot allows you to
                manually adjust application groupings to better fit your specific needs. You can match desired application instances by defining
                <a href="https://en.wikipedia.org/wiki/Glob_(programming)" target="_blank">glob patterns</a>
                for <var>instance_name</var>. Note that this is not applicable to Kubernetes applications.
            </p> -->

            <CustomApplications />
        </template>

       

        
    </div>
</template>

<script>
import ProjectSettings from './ProjectSettings.vue';
import ProjectStatus from './ProjectStatus.vue';
import ProjectDelete from './ProjectDelete.vue';
import Inspections from './Inspections.vue';
import ApplicationCategories from './ApplicationCategories.vue';
import IntegrationPrometheus from './IntegrationPrometheus.vue';
import IntegrationClickhouse from './IntegrationClickhouse.vue';
import CustomApplications from './CustomApplications.vue';

export default {
    props: {
        projectId: String,
        tab: String,
    },

    components: {
        CustomApplications,
        IntegrationPrometheus,
        IntegrationClickhouse,
        Inspections,
        ProjectSettings,
        ProjectStatus,
        ProjectDelete,
        ApplicationCategories,
    },

    mounted() {
        if (!this.tabs.find((t) => t.id === this.tab)) {
            this.$router.replace({ params: { tab: undefined } });
        }
    },

    computed: {
        tabs() {
            const disabled = !this.projectId;
            return [
                { id: undefined, name: 'General' },
                { id: 'prometheus', name: 'Prometheus', disabled },
                { id: 'clickhouse', name: 'Clickhouse', disabled },
                { id: 'inspections', name: 'Inspections', disabled },
                { id: 'applications', name: 'Applications', disabled },
                { id: 'notifications', name: 'Notifications', disabled },
            ];
        },
    },
};
</script>

<style scoped></style>
