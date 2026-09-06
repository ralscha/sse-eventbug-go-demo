import * as echarts from 'echarts/core';
import { GaugeChart } from 'echarts/charts';
import { CanvasRenderer } from 'echarts/renderers';
echarts.use([GaugeChart, CanvasRenderer]);

export default class App {
    constructor() {
        this.names = ['s1', 's2', 's3', 's4', 's5'];
        this.gauges = [];
        this.onResize = () => this.gauges.forEach((gauge) => gauge.resize());

        const threshold = [0.1, 0.2, 0.7, 0.5, 0.9];

        for (const [index, name] of this.names.entries()) {
            const chart = echarts.init(document.getElementById('chart' + (index + 1)));
            chart.setOption(this.getChartOption(name, threshold[index]));
            this.gauges.push(chart);
        }
        window.addEventListener('resize', this.onResize);
    }

    start() {
        this.stop();
        this.eventSource = new EventSource(`/register/${crypto.randomUUID()}`);
        this.eventSource.addEventListener('message', this.onMessage.bind(this), false);
        this.eventSource.addEventListener('dto', m => console.log(m));
        this.eventSource.onerror = this.onError;
        this.eventSource.onopen = this.onOpen;
    }

    stop() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }
    }

    onMessage(response) {
        try {
            this.handleResponse(JSON.parse(response.data));
        } catch (error) {
            console.error('Invalid gauge event', error);
        }
    }

    onError() {
        console.log("Error occurred");
    }

    onOpen() {
        console.log("Connection to server opened");
    }

    handleResponse(data) {
        if (!Array.isArray(data) || data.length !== this.gauges.length) {
            return;
        }
        for (const [index, gauge] of this.gauges.entries()) {
            gauge.setOption({
                series: {
                    data: [{
                        name: this.names[index],
                        value: data[index]
                    }]
                }
            });
        }
    }

    getChartOption(name, threshold) {
        return {
            series: [{
                startAngle: 180,
                endAngle: 0,
                center: ['50%', '90%'],
                radius: 100,
                min: 0,
                max: 30,
                name: 'Serie',
                type: 'gauge',
                splitNumber: 3,
                data: [{
                    value: 16,
                    name: name
                }],
                title: {
                    show: true,
                    offsetCenter: ['-100%', '-90%'],
                    color: '#333',
                    fontSize: 15
                },
                axisLine: {
                    lineStyle: {
                        color: [[threshold, '#ff4500'], [1, 'lightgreen']],
                        width: 8
                    }
                },
                axisTick: {
                    length: 11,
                    lineStyle: {
                        color: 'inherit'
                    }
                },
                splitLine: {
                    length: 15,
                    lineStyle: {
                        color: 'inherit'
                    }
                },
                detail: {
                    show: true,
                    offsetCenter: ['100%', '-100%'],
                    color: 'inherit',
                    fontSize: 25
                }

            }]
        };
    }


}
